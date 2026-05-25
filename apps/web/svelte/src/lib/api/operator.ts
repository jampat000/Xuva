import { apiClient, type ApiClient } from './client';

export interface CatalogSummaryResponse {
	libraries?: number;
	mediaSources?: number;
	movies?: number;
	series?: number;
	episodes?: number;
	scanRuns?: number;
	unprobed?: number;
}

export interface CatalogHealthResponse {
	summary?: CatalogSummaryResponse;
	/** Total disk usage of all media sources in bytes (SUM of size_bytes). */
	totalSizeBytes?: number;
	needsReview?: number;
	unprobed?: number;
	unsupported?: number;
	highBitrate?: number;
	withSubtitles?: number;
}

export interface SystemStatusResponse {
	collectedAt?: string;
	/** ISO-8601 timestamp the server process started. Used to compute uptime client-side. */
	serverStartedAt?: string;
	cpu?: {
		percent?: number;
		cores?: number;
	};
	memory?: {
		totalBytes?: number;
		availableBytes?: number;
		usedBytes?: number;
		usedPercent?: number;
	};
	process?: {
		goAllocBytes?: number;
		goSysBytes?: number;
		goroutines?: number;
	};
	network?: {
		receiveBps?: number;
		transmitBps?: number;
		linkSpeedBps?: number;
		interfaces?: Array<{
			name?: string;
			receiveBps?: number;
			transmitBps?: number;
			linkSpeedBps?: number;
		}>;
	};
	disks?: Array<{
		name?: string;
		path?: string;
		usedBytes?: number;
		usedPercent?: number;
		freeBytes?: number;
		totalBytes?: number;
		writable?: boolean;
		error?: string;
		sharedWithData?: boolean;
	}>;
}
