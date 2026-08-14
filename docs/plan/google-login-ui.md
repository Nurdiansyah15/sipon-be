# Plan: Google Login UI (sipon-ui)

## Context

Frontend SIPON perlu mendukung Google Sign-In dan Linked Accounts Management. Implementasi mengikuti pola k-forum-backoffice. Saat ini:

- Tombol Google di halaman login/register **sudah ada** tapi belum fungsional (tidak ada `@click` handler).
- `SetPasswordForm` sudah ada untuk user tanpa password (prasyarat unlink Google).
- Belum ada Linked Accounts section di profil.

## Lingkup

### Yang Dikerjakan
1. Google Sign-In button di halaman login (GSI script)
2. `loginWithGoogle` action di auth store
3. Linked Accounts tab di halaman profil
4. Link Google manual dari profil
5. Unlink Google dengan konfirmasi modal
6. Set password form (sudah ada, perlu update pesan untuk konteks Google)
7. Konfigurasi `NUXT_PUBLIC_GOOGLE_CLIENT_ID`

### Yang Tidak Dikerjakan
- Google Sign-In di halaman register (register tetap manual, Google login otomatis membuat akun)
- Apple Sign In

## Konfigurasi

### Environment Variable

**File**: `.env.example`

```bash
# Google OAuth Client ID (dari Google Cloud Console)
NUXT_PUBLIC_GOOGLE_CLIENT_ID=
```

**File**: `nuxt.config.ts`

```ts
runtimeConfig: {
  public: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8888',
    googleClientId: process.env.NUXT_PUBLIC_GOOGLE_CLIENT_ID || '',
  },
},
```

## Types

### File baru: `shared/types/UserSettings.ts`

```ts
export interface GoogleLinkedAccount {
  linked: boolean
  email: string | null
  can_unlink: boolean
}

export interface LinkedAccountsResponse {
  google: GoogleLinkedAccount
}
```

### Update: `shared/types/Auth.ts`

```ts
export interface GoogleLoginRequest {
  id_token: string
  device_id?: string
}

export interface LinkGoogleRequest {
  id_token: string
}
```

## Auth Store Changes

### File: `app/stores/auth.ts`

Tambah action `loginWithGoogle`:

```ts
async loginWithGoogle(idToken: string) {
  this.isLoading = true
  this.error = null
  try {
    const api = useApi()
    const res = await api.post<ApiSuccess<LoginResponse>>(
      '/api/v1/web/auth/login/google',
      { id_token: idToken },
    )
    this.setSession(res.data)
    await this.fetchSession()
  } catch (err) {
    this.error = parseApiError(err, 'Gagal masuk dengan Google.')
    throw err
  } finally {
    this.isLoading = false
  }
},
```

### File baru: `app/stores/userSettings.ts`

```ts
import { defineStore } from 'pinia'
import { useApi } from '~/composables/useApi'
import { parseApiError } from '~/utils/errorParser'
import type { ApiSuccess } from '#shared/types/ApiResponse'
import type { LinkedAccountsResponse } from '#shared/types/UserSettings'

export const useUserSettingsStore = defineStore('userSettings', () => {
  const api = useApi()
  const toast = useToast()

  const linkedAccounts = ref<LinkedAccountsResponse | null>(null)
  const isLoadingLinkedAccounts = ref(false)
  const isLinkingGoogle = ref(false)
  const isUnlinkingGoogle = ref(false)

  async function fetchLinkedAccounts(force = false) {
    if (!force && linkedAccounts.value) return linkedAccounts.value
    isLoadingLinkedAccounts.value = true
    try {
      const res = await api.get<ApiSuccess<LinkedAccountsResponse>>(
        '/api/v1/web/auth/linked-accounts',
      )
      linkedAccounts.value = res.data
      return res.data
    } finally {
      isLoadingLinkedAccounts.value = false
    }
  }

  async function linkGoogle(idToken: string) {
    isLinkingGoogle.value = true
    try {
      await api.post('/api/v1/web/auth/linked-accounts/google', {
        id_token: idToken,
      })
      await fetchLinkedAccounts(true)
      toast.add({ title: 'Akun Google berhasil ditautkan', color: 'success' })
    } catch (err) {
      toast.add({
        title: 'Gagal menautkan akun Google',
        description: parseApiError(err, 'Terjadi kesalahan'),
        color: 'error',
      })
      throw err
    } finally {
      isLinkingGoogle.value = false
    }
  }

  async function unlinkGoogle() {
    isUnlinkingGoogle.value = true
    try {
      await api.delete('/api/v1/web/auth/linked-accounts/google')
      if (linkedAccounts.value) {
        linkedAccounts.value.google = {
          linked: false,
          email: null,
          can_unlink: false,
        }
      }
      toast.add({ title: 'Akun Google berhasil dilepas', color: 'success' })
    } catch (err) {
      toast.add({
        title: 'Gagal melepas akun Google',
        description: parseApiError(err, 'Terjadi kesalahan'),
        color: 'error',
      })
      throw err
    } finally {
      isUnlinkingGoogle.value = false
    }
  }

  return {
    linkedAccounts,
    isLoadingLinkedAccounts,
    isLinkingGoogle,
    isUnlinkingGoogle,
    fetchLinkedAccounts,
    linkGoogle,
    unlinkGoogle,
  }
})
```

## Halaman Login

### Update: `app/pages/auth/login.vue`

**Perubahan utama:**

1. Load Google GSI script via `useHead()`
2. Init `google.accounts.id.initialize()` di `onMounted`
3. Render Google button ke container `#google-btn-container`
4. Callback `handleCredentialResponse` -> `authStore.loginWithGoogle(idToken)`
5. Ganti tombol Google statis dengan GSI container + fallback button

```vue
<script setup lang="ts">
// ... existing code ...

const config = useRuntimeConfig()
const clientId = config.public.googleClientId
const googleGsiLoaded = ref(false)

function initGoogleGsi() {
  if (typeof window === 'undefined' || !(window as any).google) {
    setTimeout(initGoogleGsi, 100)
    return
  }
  googleGsiLoaded.value = true
  const google = (window as any).google
  google.accounts.id.initialize({
    client_id: clientId,
    callback: handleCredentialResponse,
  })
  const container = document.getElementById('google-btn-container')
  if (container) {
    google.accounts.id.renderButton(container, {
      theme: 'outline',
      size: 'large',
      width: '380',
      type: 'standard',
    })
  }
}

async function handleCredentialResponse(response: any) {
  const idToken = response.credential
  try {
    await authStore.loginWithGoogle(idToken)
    toast.add({ title: 'Berhasil masuk dengan Google', color: 'success' })
    const redirect = (route.query.redirect as string) || '/dashboard'
    await navigateTo(redirect)
  } catch {
    toast.add({
      title: 'Gagal masuk dengan Google',
      description: authStore.error ?? undefined,
      color: 'error',
    })
  }
}

onMounted(() => {
  initGoogleGsi()
})

useHead({
  script: [
    { src: 'https://accounts.google.com/gsi/client', async: true, defer: true },
  ],
})
</script>
```

**Template update** (ganti bagian `<UButton>` Google statis):

```vue
<!-- Google Sign-In Container -->
<div v-show="googleGsiLoaded" class="flex justify-center w-full min-h-[40px]">
  <div id="google-btn-container" class="w-full flex justify-center"></div>
</div>

<!-- Fallback: tombol Google manual jika GSI belum load -->
<UButton
  v-show="!googleGsiLoaded"
  variant="outline"
  block
  size="lg"
  class="text-gray-700 dark:text-gray-300"
  :loading="authStore.isLoading"
  @click="initGoogleGsi()"
>
  <svg class="h-5 w-5" viewBox="0 0 24 24">
    <!-- ... Google SVG icon (sama seperti sekarang) ... -->
  </svg>
  Google
</UButton>
```

## Komponen Profil Baru

### File baru: `app/components/profile/LinkedAccountsPanel.vue`

```vue
<script setup lang="ts">
import { useUserSettingsStore } from '~/stores/userSettings'

const store = useUserSettingsStore()
const isUnlinkModalOpen = ref(false)

onMounted(async () => {
  await store.fetchLinkedAccounts()
})

function confirmUnlink() {
  store.unlinkGoogle()
  isUnlinkModalOpen.value = false
}
</script>

<template>
  <div>
    <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">
      Akun Tertaut
    </h3>
    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
      Kelola akun eksternal yang terhubung dengan akun SIPON Anda.
    </p>

    <!-- Google Row -->
    <div class="mt-4 flex flex-wrap items-center justify-between gap-4 rounded-lg border border-gray-200 p-4 dark:border-gray-700/50">
      <div class="flex items-center gap-3">
        <!-- Google icon SVG -->
        <svg class="h-6 w-6" viewBox="0 0 24 24">
          <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
          <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
          <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
          <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
        </svg>
        <div>
          <p class="text-sm font-semibold text-gray-900 dark:text-gray-100">Google</p>
          <template v-if="store.linkedAccounts?.google.linked">
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ store.linkedAccounts.google.email }}
            </p>
          </template>
          <p v-else class="text-xs text-gray-400 dark:text-gray-500">
            Belum tertaut
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <UBadge
          :color="store.linkedAccounts?.google.linked ? 'success' : 'neutral'"
          variant="subtle"
          size="sm"
        >
          {{ store.linkedAccounts?.google.linked ? 'Tertaut' : 'Belum tertaut' }}
        </UBadge>

        <template v-if="store.linkedAccounts?.google.linked">
          <UTooltip
            v-if="!store.linkedAccounts.google.can_unlink"
            text="Atur kata sandi terlebih dahulu (tab Keamanan) sebelum melepas Google."
          >
            <UButton
              variant="soft"
              size="xs"
              color="error"
              disabled
            >
              Lepas
            </UButton>
          </UTooltip>
          <UButton
            v-else
            variant="soft"
            size="xs"
            color="error"
            :loading="store.isUnlinkingGoogle"
            @click="isUnlinkModalOpen = true"
          >
            Lepas
          </UButton>
        </template>
      </div>
    </div>

    <!-- Warning jika tidak bisa unlink -->
    <UAlert
      v-if="store.linkedAccounts?.google.linked && !store.linkedAccounts.google.can_unlink"
      class="mt-4"
      color="warning"
      variant="subtle"
      title="Atur kata sandi terlebih dahulu"
      description="Anda perlu mengatur kata sandi di tab Keamanan sebelum bisa melepas akun Google."
    />

    <!-- Unlink Confirmation Modal -->
    <UModal v-model:open="isUnlinkModalOpen" title="Lepas Akun Google" :ui="{ content: 'max-w-sm' }">
      <template #body>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Anda tidak akan bisa masuk dengan Google lagi. Anda masih bisa masuk dengan
          email/username dan kata sandi.
        </p>
      </template>
      <template #footer>
        <div class="flex gap-2 justify-end">
          <UButton variant="outline" @click="isUnlinkModalOpen = false">
            Batal
          </UButton>
          <UButton color="error" :loading="store.isUnlinkingGoogle" @click="confirmUnlink">
            Lepas
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
```

## Update Halaman Profil

### Update: `app/pages/profile/index.vue`

Tambah tab baru "Akun Tertaut":

```vue
const tabItems: TabsItem[] = [
  { label: 'Informasi Akun', icon: 'i-lucide-user', value: 'account' },
  { label: 'Akun Tertaut', icon: 'i-lucide-link', value: 'linked' },      // BARU
  { label: 'Data Santri', icon: 'i-lucide-graduation-cap', value: 'santri' },
  { label: 'Roles & Permissions', icon: 'i-lucide-shield', value: 'roles' },
  { label: 'Keamanan', icon: 'i-lucide-lock', value: 'security' },
]
```

Tambah render component di template:

```vue
<ProfileAccountInfoPanel v-if="item.value === 'account'" />
<ProfileLinkedAccountsPanel v-else-if="item.value === 'linked'" />   <!-- BARU -->
<ProfileSantriProfilePanel v-else-if="item.value === 'santri'" />
<ProfileRolesPermissionsPanel v-else-if="item.value === 'roles'" />
```

### Update: `app/components/profile/SetPasswordForm.vue`

Tambah info banner jika user login via Google:

```vue
<template>
  <div class="max-w-md">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Atur Kata Sandi</h3>
    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
      Akun Anda belum memiliki kata sandi. Atur kata sandi untuk bisa masuk menggunakan
      email/username — ini juga diperlukan sebelum bisa melepas akun Google dari
      tab "Akun Tertaut".
    </p>
    <!-- ... form tetap sama ... -->
  </div>
</template>
```

## Struktur File yang Berubah/Dibuat

```
sipon-ui/
  nuxt.config.ts                              # UPDATE: tambah googleClientId
  .env.example                                # UPDATE: tambah NUXT_PUBLIC_GOOGLE_CLIENT_ID
  shared/types/
    UserSettings.ts                           # BARU: LinkedAccounts types
    Auth.ts                                   # UPDATE: tambah GoogleLoginRequest, LinkGoogleRequest
  app/
    stores/
      auth.ts                                 # UPDATE: tambah loginWithGoogle
      userSettings.ts                         # BARU: linked accounts store
    pages/
      auth/
        login.vue                             # UPDATE: GSI script + Google button handler
      profile/
        index.vue                             # UPDATE: tambah tab "Akun Tertaut"
    components/
      profile/
        LinkedAccountsPanel.vue               # BARU: linked accounts UI
        SetPasswordForm.vue                   # UPDATE: info text untuk konteks Google
```

## Data Flow

```
LOGIN FLOW:
  [Google GSI Script] -> [Google OAuth popup] -> idToken
    -> handleCredentialResponse()
      -> authStore.loginWithGoogle(idToken)
        -> POST /api/v1/web/auth/login/google { id_token }
          -> Response { token, refresh_token, user }
            -> setSession() + fetchSession()
              -> navigateTo('/dashboard')

LINKED ACCOUNTS FLOW:
  [ProfileLinkedAccountsPanel onMounted]
    -> userSettingsStore.fetchLinkedAccounts()
      -> GET /api/v1/web/auth/linked-accounts
        -> Response { google: { linked, email, can_unlink } }

UNLINK FLOW:
  [User klik "Lepas"] -> [UModal konfirmasi]
    -> userSettingsStore.unlinkGoogle()
      -> DELETE /api/v1/web/auth/linked-accounts/google
        -> Optimistic update: linked=false, email=null

SET PASSWORD FLOW (prasyarat unlink):
  [SetPasswordForm] -> authStore.setPassword({ new_password })
    -> POST /api/v1/web/auth/set-password
      -> user.has_password = true
```

## Error Handling

| Error Code dari API | Penanganan UI |
|---------------------|---------------|
| `GOOGLE_TOKEN_INVALID` | Toast: "Token Google tidak valid. Silakan coba lagi." |
| `GOOGLE_EMAIL_NOT_VERIFIED` | Toast: "Email Google belum terverifikasi." |
| `GOOGLE_SUB_ALREADY_USED` | Toast: "Akun Google sudah terhubung ke pengguna lain." |
| `GOOGLE_ALREADY_LINKED` | Toast: "Akun Google sudah tertaut." |
| `GOOGLE_UNLINK_REQUIRES_PASSWORD` | Toast: "Atur kata sandi terlebih dahulu." + arahkan ke tab Keamanan |
| `GOOGLE_NOT_LINKED` | Toast: "Akun Google tidak tertaut." |

## Checklist Implementasi

- [ ] Tambah `NUXT_PUBLIC_GOOGLE_CLIENT_ID` di `.env.example` dan `nuxt.config.ts`
- [ ] Buat `shared/types/UserSettings.ts`
- [ ] Update `shared/types/Auth.ts` dengan Google types
- [ ] Tambah `loginWithGoogle` action di auth store
- [ ] Buat `app/stores/userSettings.ts`
- [ ] Update `app/pages/auth/login.vue` dengan GSI script
- [ ] Buat `app/components/profile/LinkedAccountsPanel.vue`
- [ ] Update `app/pages/profile/index.vue` tambah tab
- [ ] Update `app/components/profile/SetPasswordForm.vue` info text
- [ ] Test flow: login Google -> linked accounts -> unlink -> set password -> unlink lagi
