package kesantrian

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	dokumenAset "sipon-be/internal/modules/dokumen_aset"
	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/kesantrian/application/command"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	"sipon-be/internal/modules/kesantrian/application/query"
	"sipon-be/internal/modules/kesantrian/domain/surat/service"
	"sipon-be/internal/modules/kesantrian/infrastructure/dokumenasetgateway"
	"sipon-be/internal/modules/kesantrian/infrastructure/external"
	"sipon-be/internal/modules/kesantrian/infrastructure/identitygateway"
	"sipon-be/internal/modules/kesantrian/infrastructure/persistence"
	"sipon-be/internal/modules/kesantrian/infrastructure/scopegateway"
	kesantrianHTTP "sipon-be/internal/modules/kesantrian/interfaces/http"
	"sipon-be/internal/shared/config"
	messaging "sipon-be/internal/modules/messaging"
)

// Module's exported surface is method-only, zero exported fields — mirrors
// identity.Module. Only RegisterRoutes exists here (no RateLimiter/Contract
// needed yet: kesantrian has no rate limiter of its own and no other module
// calls into it — YAGNI, see docs/architecture/module-boundaries.md).
type Module struct {
	handler                       *kesantrianHTTP.SantriHandler
	persuratanHandler             *kesantrianHTTP.PersuratanHandler
	createSantriFromPendaftaranUC *command.CreateSantriFromPendaftaranUseCase
	createSantriUC                *command.CreateSantriUseCase
	approveSantriRequestUC        *command.ApproveSantriRequestUseCase
	listActiveSantriIDsUC         *query.ListActiveSantriIDsUseCase
	getSantriByUserIDUC           *query.GetSantriByUserIDUseCase
	getSantriByIDUC               *query.GetSantriByIDUseCase
	getSantriByNISUC              *query.GetSantriByNISUseCase
	listActiveSantriWithUserIDUC  *query.ListActiveSantriWithUserIDUseCase
	fileUploader                  ports.FileUploader
	provisioner                   ports.AccountProvisioner
	jwtAuth                       gin.HandlerFunc
	principalLoad                 gin.HandlerFunc
}

// NewModule takes identity as identity.Contract (not *identity.Module) —
// this is the enforcement point that keeps kesantrian from ever reaching
// identity's RegisterRoutes/RateLimiter/domain/application internals. jwtAuth
// and principalLoad are the two ready-made middleware funcs sourced from
// identity.Module.AuthMiddleware()/PrincipalMiddleware() in cmd/api/main.go
// — see docs/architecture/module-boundaries.md and the kesantrian port plan.
func NewModule(
	db *sql.DB,
	redisClient *redis.Client,
	cfg *config.Config,
	identityContract identity.Contract,
	dokumenAsetContract dokumenAset.Contract,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	santriRepo := persistence.NewPostgresSantriRepository(db)
	dokumenRepo := persistence.NewPostgresSantriDokumenRepository(db)
	requestRepo := persistence.NewPostgresSantriRequestRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	provisioner := identitygateway.New(identityContract)
	scopeReader := scopegateway.New(identityContract)

	fileUploader, _ := external.NewMinioFileUploader(
		cfg.Minio.Endpoint,
		cfg.Minio.PublicEndpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.PrivateBucket,
		cfg.Minio.UseSSL,
		cfg.Minio.InternalUseSSL,
	)

	getSantriUC := query.NewGetSantriUseCase(santriRepo, provisioner, fileUploader)
	getSantriDetailUC := query.NewGetSantriDetailUseCase(santriRepo, provisioner, fileUploader)
	updateSantriUC := command.NewUpdateSantriUseCase(santriRepo, provisioner)
	requestSantriUC := command.NewRequestSantriUseCase(santriRepo, requestRepo, transactor)

	createSantriUC := command.NewCreateSantriUseCase(santriRepo, provisioner, transactor)
	importSantriUC := command.NewImportSantriUseCase(santriRepo, provisioner, transactor)
	listSantriUC := query.NewListSantriUseCase(santriRepo, provisioner, scopeReader)
	listSantriRequestsUC := query.NewListSantriRequestsUseCase(requestRepo, provisioner)
	approveSantriRequestUC := command.NewApproveSantriRequestUseCase(requestRepo, santriRepo, provisioner, transactor)
	rejectSantriRequestUC := command.NewRejectSantriRequestUseCase(requestRepo, transactor)

	dokumenPresignUC := command.NewDokumenPresignUseCase(fileUploader)
	dokumenConfirmUC := command.NewDokumenConfirmUseCase(santriRepo, dokumenRepo, fileUploader, transactor)
	dokumenListUC := query.NewDokumenListUseCase(santriRepo, dokumenRepo)
	adminDokumenListUC := query.NewAdminDokumenListUseCase(santriRepo, dokumenRepo)
	dokumenAccessUC := query.NewDokumenAccessUseCase(santriRepo, dokumenRepo, fileUploader)
	dokumenDeleteUC := command.NewDokumenDeleteUseCase(santriRepo, dokumenRepo, fileUploader, transactor)
	dokumenVerifyUC := command.NewDokumenVerifyUseCase(dokumenRepo, transactor)
	dokumenRejectUC := command.NewDokumenRejectUseCase(dokumenRepo, transactor)

	createSantriFromPendaftaranUC := command.NewCreateSantriFromPendaftaranUseCase(santriRepo, dokumenRepo, provisioner, transactor)
	changeSantriStatusUC := command.NewChangeSantriStatusUseCase(santriRepo)
	listActiveSantriIDsUC := query.NewListActiveSantriIDsUseCase(santriRepo)
	getSantriByUserIDUC := query.NewGetSantriByUserIDUseCase(santriRepo)
	getSantriByIDUC := query.NewGetSantriByIDUseCase(santriRepo)
	getSantriByNISUC := query.NewGetSantriByNISUseCase(santriRepo)
	listActiveSantriWithUserIDUC := query.NewListActiveSantriWithUserIDUseCase(santriRepo)

	tipeSuratRepo := persistence.NewPostgresTipeSuratRepository(db)
	suratRepo := persistence.NewPostgresSuratRepository(db)
	nomorGenerator := service.NewNomorGenerator(suratRepo)
	dokumenAsetReader := dokumenasetgateway.New(dokumenAsetContract)

	createTipeSuratUC := command.NewCreateTipeSuratUseCase(tipeSuratRepo)
	updateTipeSuratUC := command.NewUpdateTipeSuratUseCase(tipeSuratRepo)
	deleteTipeSuratUC := command.NewDeleteTipeSuratUseCase(tipeSuratRepo)
	listTipeSuratUC := query.NewListTipeSuratUseCase(tipeSuratRepo)
	getTipeSuratUC := query.NewGetTipeSuratUseCase(tipeSuratRepo)
	createSuratUC := command.NewCreateSuratUseCase(suratRepo, tipeSuratRepo, nomorGenerator, transactor, scopeReader)
	deleteSuratUC := command.NewDeleteSuratUseCase(suratRepo, scopeReader)
	addSuratDokumenUC := command.NewAddSuratDokumenUseCase(suratRepo, scopeReader)
	removeSuratDokumenUC := command.NewRemoveSuratDokumenUseCase(suratRepo, scopeReader)
	listSuratUC := query.NewListSuratUseCase(suratRepo, scopeReader)
	getSuratUC := query.NewGetSuratUseCase(suratRepo, tipeSuratRepo, scopeReader)
	getSuratDownloadUC := query.NewGetSuratDownloadUseCase(suratRepo, dokumenAsetReader, scopeReader)

	persuratanHandler := kesantrianHTTP.NewPersuratanHandler(
		createTipeSuratUC, updateTipeSuratUC, deleteTipeSuratUC,
		listTipeSuratUC, getTipeSuratUC,
		createSuratUC, deleteSuratUC,
		addSuratDokumenUC, removeSuratDokumenUC,
		listSuratUC, getSuratUC, getSuratDownloadUC,
	)

	handler := kesantrianHTTP.NewSantriHandler(
		getSantriUC,
		getSantriDetailUC,
		updateSantriUC,
		requestSantriUC,
		createSantriUC,
		importSantriUC,
		listSantriUC,
		listSantriRequestsUC,
		approveSantriRequestUC,
		rejectSantriRequestUC,
		dokumenPresignUC,
		dokumenConfirmUC,
		dokumenListUC,
		adminDokumenListUC,
		dokumenAccessUC,
		dokumenDeleteUC,
		dokumenVerifyUC,
		dokumenRejectUC,
		createSantriFromPendaftaranUC,
		changeSantriStatusUC,
	)

	return &Module{handler: handler, persuratanHandler: persuratanHandler, createSantriFromPendaftaranUC: createSantriFromPendaftaranUC, createSantriUC: createSantriUC, approveSantriRequestUC: approveSantriRequestUC, listActiveSantriIDsUC: listActiveSantriIDsUC, getSantriByUserIDUC: getSantriByUserIDUC, getSantriByIDUC: getSantriByIDUC, getSantriByNISUC: getSantriByNISUC, listActiveSantriWithUserIDUC: listActiveSantriWithUserIDUC, fileUploader: fileUploader, provisioner: provisioner, jwtAuth: jwtAuth, principalLoad: principalLoad}
}

// SetAkademikProvisioner late-binds the akademik port. Called in
// cmd/api/main.go after the akademik module is constructed, because akademik
// depends on kesantrian (contract) at construction time — passing it in the
// constructor would create an import cycle.
func (m *Module) SetAkademikProvisioner(p ports.AkademikProvisioner) {
	m.createSantriUC.SetAkademikProvisioner(p)
	m.createSantriFromPendaftaranUC.SetAkademikProvisioner(p)
	m.approveSantriRequestUC.SetAkademikProvisioner(p)
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	kesantrianHTTP.RegisterRoutes(grp, m.handler, m.persuratanHandler, m.jwtAuth, m.principalLoad)
}

func (m *Module) CreateSantriFromPendaftaran(ctx context.Context, in CreateSantriFromPendaftaranInput) (*CreateSantriFromPendaftaranResult, error) {
	dok := make([]command.PendaftaranDokumenInput, len(in.Dokumen))
	for i, d := range in.Dokumen {
		dok[i] = command.PendaftaranDokumenInput{
			Kind:             d.Kind,
			Key:              d.Key,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			Size:             d.Size,
			VerifiedBy:       d.VerifiedBy,
			VerifiedAt:       d.VerifiedAt,
		}
	}
	result, err := m.createSantriFromPendaftaranUC.Execute(ctx, command.CreateSantriFromPendaftaranCmd{
		UserID:    in.UserID,
		Gender:    in.Gender,
		EntryYear: in.EntryYear,

		ProgramID:       in.ProgramID,
		Nickname:        in.Nickname,
		Program:         in.Program,
		Hobby:           in.Hobby,
		Purpose:         in.Purpose,
		MotivationEntry: in.MotivationEntry,
		POB:             in.POB,
		DOB:             in.DOB,
		Blood:           in.Blood,

		Address:     in.Address,
		SubDistrict: in.SubDistrict,
		District:    in.District,
		Province:    in.Province,
		PostalCode:  in.PostalCode,

		PreviousPondokName:    in.PreviousPondokName,
		PreviousPondokAddress: in.PreviousPondokAddress,
		PreviousPondokDiv:     in.PreviousPondokDiv,
		PreviousPondokTime:    in.PreviousPondokTime,

		NIK:   in.NIK,
		NoKK:  in.NoKK,
		NISN:  in.NISN,
		NoKIP: in.NoKIP,
		NoKKS: in.NoKKS,
		NoPKH: in.NoPKH,

		Workplace:  in.Workplace,
		Department: in.Department,

		HomeStatus: in.HomeStatus,

		Father:         in.Father,
		FatherPN:       in.FatherPN,
		FatherNIK:      in.FatherNIK,
		FatherJob:      in.FatherJob,
		FatherGraduate: in.FatherGraduate,
		FatherIncome:   in.FatherIncome,

		Mother:         in.Mother,
		MotherPN:       in.MotherPN,
		MotherNIK:      in.MotherNIK,
		MotherJob:      in.MotherJob,
		MotherGraduate: in.MotherGraduate,
		MotherIncome:   in.MotherIncome,

		GuardianRelationship: in.GuardianRelationship,
		Guardian:             in.Guardian,
		GuardianPN:           in.GuardianPN,
		GuardianNIK:          in.GuardianNIK,
		GuardianJob:          in.GuardianJob,
		GuardianGraduate:     in.GuardianGraduate,
		GuardianIncome:       in.GuardianIncome,

		Dokumen: dok,
	})
	if err != nil {
		return nil, err
	}
	return &CreateSantriFromPendaftaranResult{
		SantriID: result.SantriID,
		NIS:      result.NIS,
	}, nil
}

func (m *Module) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	return m.fileUploader.EnsurePendingUploadLifecycle(ctx, expireDays)
}

func (m *Module) ListActiveSantriIDs(ctx context.Context) ([]string, error) {
	return m.listActiveSantriIDsUC.Execute(ctx)
}

func (m *Module) GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error) {
	result, err := m.getSantriByUserIDUC.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	return m.enrichBasicInfo(ctx, &SantriBasicInfo{
		SantriID: result.SantriID,
		UserID:   result.UserID,
		NIS:      result.NIS,
		Status:   result.Status,
	}), nil
}

func (m *Module) GetSantriByID(ctx context.Context, santriID string) (*SantriBasicInfo, error) {
	result, err := m.getSantriByIDUC.Execute(ctx, santriID)
	if err != nil {
		return nil, err
	}
	return m.enrichBasicInfo(ctx, &SantriBasicInfo{
		SantriID: result.SantriID,
		UserID:   result.UserID,
		NIS:      result.NIS,
		Status:   result.Status,
	}), nil
}

func (m *Module) GetSantriByNIS(ctx context.Context, nis string) (*SantriBasicInfo, error) {
	result, err := m.getSantriByNISUC.Execute(ctx, nis)
	if err != nil {
		return nil, err
	}
	return m.enrichBasicInfo(ctx, &SantriBasicInfo{
		SantriID: result.SantriID,
		UserID:   result.UserID,
		NIS:      result.NIS,
		Status:   result.Status,
	}), nil
}

// enrichBasicInfo best-effort attaches the user's fullname (from identity) to a
// SantriBasicInfo. Enrichment failure is logged, not fatal — mirroring the
// N+1-by-design pattern used in query/list_santri.go.
func (m *Module) enrichBasicInfo(ctx context.Context, info *SantriBasicInfo) *SantriBasicInfo {
	if info == nil || info.UserID == "" {
		return info
	}
	summary, err := m.provisioner.GetUserSummary(ctx, info.UserID)
	if err != nil {
		slog.Warn("kesantrian: user summary enrichment failed", "user_id", info.UserID, "error", err)
		return info
	}
	info.Fullname = summary.Fullname
	return info
}

func (m *Module) ListActiveSantriWithUserID(ctx context.Context) ([]SantriBasicInfo, error) {
	results, err := m.listActiveSantriWithUserIDUC.Execute(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]SantriBasicInfo, len(results))
	for i, r := range results {
		infos[i] = SantriBasicInfo{
			SantriID: r.SantriID,
			UserID:   r.UserID,
			NIS:      r.NIS,
			Status:   r.Status,
		}
	}
	return infos, nil
}

func (m *Module) RegisterMessageHandlers(_ messaging.Contract) ([]messaging.Binding, error) {
	return nil, nil
}
