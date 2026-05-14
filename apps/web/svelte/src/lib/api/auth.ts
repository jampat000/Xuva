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
	profiles?: Array<{
		id?: string;
		name?: string;
	}>;
	[key: string]: unknown;
}

export interface LoginRequest {
	username: string;
	password: string;
}

export interface BootstrapRequest extends LoginRequest {
	displayName?: string;
}

export interface UpdatePasswordRequest {
	password: string;
}

export interface UserAccount {
	id: string;
	username: string;
	displayName: string;
	role: string;
	createdAt?: string;
}

export interface ListUsersResponse {
	users?: UserAccount[];
}

export interface CreateUserRequest {
	username: string;
	password: string;
	displayName?: string;
	role: 'admin' | 'standard';
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

export async function getUsers(client: ApiClient = apiClient): Promise<ListUsersResponse> {
	return client.request<ListUsersResponse>('/api/users');
}

export async function createUser(
	payload: CreateUserRequest,
	client: ApiClient = apiClient
): Promise<{ user: UserAccount }> {
	return client.send<{ user: UserAccount }, CreateUserRequest>('/api/users', payload, 'POST');
}

export async function deleteUser(
	userID: string,
	client: ApiClient = apiClient
): Promise<{ status: string }> {
	return client.send<{ status: string }, Record<string, never>>(
		`/api/users/${encodeURIComponent(userID)}`,
		{},
		'DELETE'
	);
}

export async function updateUserPassword(
	userID: string,
	password: string,
	client: ApiClient = apiClient
): Promise<{ status: string }> {
	return client.send<{ status: string }, UpdatePasswordRequest>(
		`/api/users/${encodeURIComponent(userID)}/password`,
		{ password },
		'POST'
	);
}
