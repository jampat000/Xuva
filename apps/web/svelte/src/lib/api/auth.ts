import { apiClient, type ApiClient } from './client';

export interface AuthSessionUser {
	id: string;
	username: string;
	displayName: string;
	role: string;
}

export interface AuthSession {
	id: string;
	expiresAt: string;
}

export interface AuthSessionResponse {
	authDisabled?: boolean;
	user?: AuthSessionUser;
	session?: AuthSession;
	csrfToken?: string;
	error?: string;
	[key: string]: unknown;
}

export interface ClientBootstrapAuthInfo {
	required?: boolean;
	bootstrapAllowed?: boolean;
	defaultUsername?: string;
	bootstrapEndpoint?: string;
}

export interface ClientBootstrapResponse {
	auth?: ClientBootstrapAuthInfo;
	[key: string]: unknown;
}

export interface LoginRequest {
	username: string;
	password: string;
}

export interface BootstrapRequest extends LoginRequest {
	displayName?: string;
}

export async function getAuthSession(client: ApiClient = apiClient): Promise<AuthSessionResponse> {
	return client.request<AuthSessionResponse>('/api/auth/session');
}

export async function getClientBootstrap(client: ApiClient = apiClient): Promise<ClientBootstrapResponse> {
	return client.request<ClientBootstrapResponse>('/api/client/bootstrap');
}

export async function login(
	payload: LoginRequest,
	client: ApiClient = apiClient
): Promise<AuthSessionResponse> {
	return client.send<AuthSessionResponse, LoginRequest>('/api/auth/login', payload, 'POST');
}

export async function bootstrapAccount(
	payload: BootstrapRequest,
	client: ApiClient = apiClient
): Promise<AuthSessionResponse> {
	return client.send<AuthSessionResponse, BootstrapRequest>('/api/auth/bootstrap', payload, 'POST');
}

export async function logout(client: ApiClient = apiClient): Promise<{ status: string }> {
	return client.send<{ status: string }, Record<string, never>>('/api/auth/logout', {}, 'POST');
}
