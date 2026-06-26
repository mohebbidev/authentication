import path from 'path'
import * as grpc from '@grpc/grpc-js'
import * as protoLoader from '@grpc/proto-loader'

const PROTO_PATH = path.resolve(__dirname, '../../proto/auth/v1/auth.proto')

const packageDef = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true,
})

const proto = grpc.loadPackageDefinition(packageDef) as any

export interface RegisterRequest { full_name: string; email: string; password: string; date_of_birth: string }
export interface RegisterResponse { user_id: string }
export interface LoginRequest { email: string; password: string }
export interface LoginResponse { access_token: string; session_id: string; session_expiry: string }
export interface RefreshTokenRequest { session_id: string }
export interface RefreshTokenResponse { access_token: string }
export interface LogoutRequest { session_id: string }
export interface LogoutResponse { }
export interface VerifyEmailRequest { token: string }
export interface VerifyEmailResponse { }
export interface RequestPasswordResetRequest { email: string }
export interface RequestPasswordResetResponse { raw_token: string }
export interface ResetPasswordRequest { token: string; new_password: string }
export interface ResetPasswordResponse { }

function call<Req, Res>(client: grpc.Client, method: string, request: Req): Promise<Res> {
    return new Promise((resolve, reject) => {
        ; (client as any)[method](request, (err: grpc.ServiceError | null, res: Res) => {
            if (err) reject(err)
            else resolve(res)
        })
    })
}

function makeAuthClient(address: string) {
    const raw = new proto.auth.AuthService(
        address,
        grpc.credentials.createInsecure(),
    )

    return {
        register: (r: RegisterRequest) => call<RegisterRequest, RegisterResponse>(raw, 'Register', r),
        login: (r: LoginRequest) => call<LoginRequest, LoginResponse>(raw, 'Login', r),
        refreshToken: (r: RefreshTokenRequest) => call<RefreshTokenRequest, RefreshTokenResponse>(raw, 'RefreshToken', r),
        logout: (r: LogoutRequest) => call<LogoutRequest, LogoutResponse>(raw, 'Logout', r),
        verifyEmail: (r: VerifyEmailRequest) => call<VerifyEmailRequest, VerifyEmailResponse>(raw, 'VerifyEmail', r),
        requestPasswordReset: (r: RequestPasswordResetRequest) => call<RequestPasswordResetRequest, RequestPasswordResetResponse>(raw, 'RequestPasswordReset', r),
        resetPassword: (r: ResetPasswordRequest) => call<ResetPasswordRequest, ResetPasswordResponse>(raw, 'ResetPassword', r),
    }
}

export type AuthClient = ReturnType<typeof makeAuthClient>

let _client: AuthClient | null = null

export function getAuthClient(address: string): AuthClient {
    if (!_client) _client = makeAuthClient(address)
    return _client
}