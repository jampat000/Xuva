export interface XuvaEventPayloadMap {
	ready: { status: string };
	[type: string]: unknown;
}

export interface EventStreamReconnectPolicy {
	initialDelayMs: number;
	maxDelayMs: number;
	backoffMultiplier: number;
}

export interface EventStreamOptions<TEvents extends Record<string, unknown>> {
	url?: string;
	withCredentials?: boolean;
	reconnect?: Partial<EventStreamReconnectPolicy>;
	eventSourceFactory?: (url: string, init?: EventSourceInit) => EventSource;
}

type TypedHandler<TEvents extends Record<string, unknown>, K extends keyof TEvents> = (
	payload: TEvents[K]
) => void;

type AnyHandler = (event: { type: string; payload: unknown }) => void;

const DEFAULT_RECONNECT_POLICY: EventStreamReconnectPolicy = {
	initialDelayMs: 800,
	maxDelayMs: 20_000,
	backoffMultiplier: 1.8
};

export class EventStream<TEvents extends Record<string, unknown> = XuvaEventPayloadMap> {
	private readonly url: string;
	private readonly withCredentials: boolean;
	private readonly reconnect: EventStreamReconnectPolicy;
	private readonly sourceFactory: (url: string, init?: EventSourceInit) => EventSource;

	private source: EventSource | null = null;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private reconnectAttempt = 0;
	private shouldReconnect = true;
	private anyHandlers = new Set<AnyHandler>();
	private typedHandlers = new Map<string, Set<(payload: unknown) => void>>();

	constructor({
		url = '/api/events',
		withCredentials = true,
		reconnect = {},
		eventSourceFactory
	}: EventStreamOptions<TEvents> = {}) {
		this.url = url;
		this.withCredentials = withCredentials;
		this.reconnect = { ...DEFAULT_RECONNECT_POLICY, ...reconnect };
		this.sourceFactory =
			eventSourceFactory ?? ((sourceUrl, init) => new EventSource(sourceUrl, init));
	}

	connect(): void {
		if (typeof window === 'undefined' || typeof EventSource === 'undefined') return;
		this.shouldReconnect = true;
		this.clearReconnectTimer();
		this.openConnection();
	}

	disconnect(): void {
		this.shouldReconnect = false;
		this.reconnectAttempt = 0;
		this.clearReconnectTimer();
		this.source?.close();
		this.source = null;
	}

	subscribe<K extends keyof TEvents & string>(eventType: K, handler: TypedHandler<TEvents, K>): () => void {
		const set = this.typedHandlers.get(eventType) ?? new Set<(payload: unknown) => void>();
		const wrappedHandler = handler as unknown as (payload: unknown) => void;
		set.add(wrappedHandler);
		this.typedHandlers.set(eventType, set);
		return () => {
			const current = this.typedHandlers.get(eventType);
			if (!current) return;
			current.delete(wrappedHandler);
			if (current.size === 0) this.typedHandlers.delete(eventType);
		};
	}

	subscribeAny(handler: AnyHandler): () => void {
		this.anyHandlers.add(handler);
		return () => this.anyHandlers.delete(handler);
	}

	private openConnection(): void {
		this.source?.close();
		const source = this.sourceFactory(this.url, { withCredentials: this.withCredentials });
		this.source = source;

		source.onopen = () => {
			this.reconnectAttempt = 0;
		};

		source.onerror = () => {
			source.close();
			this.source = null;
			if (this.shouldReconnect) this.scheduleReconnect();
		};

		source.onmessage = (event) => {
			this.emit('message', parseEventPayload(event.data));
		};

		source.addEventListener('ready', (event) => {
			if (!(event instanceof MessageEvent)) return;
			this.emit('ready', parseEventPayload(event.data));
		});
	}

	private scheduleReconnect(): void {
		this.clearReconnectTimer();
		const delay = Math.min(
			this.reconnect.maxDelayMs,
			Math.round(
				this.reconnect.initialDelayMs * this.reconnect.backoffMultiplier ** this.reconnectAttempt
			)
		);
		this.reconnectAttempt += 1;
		this.reconnectTimer = setTimeout(() => {
			this.reconnectTimer = null;
			if (!this.shouldReconnect) return;
			this.openConnection();
		}, delay);
	}

	private clearReconnectTimer(): void {
		if (!this.reconnectTimer) return;
		clearTimeout(this.reconnectTimer);
		this.reconnectTimer = null;
	}

	private emit(type: string, payload: unknown): void {
		const handlers = this.typedHandlers.get(type);
		if (handlers) {
			for (const handler of handlers) handler(payload);
		}
		for (const handler of this.anyHandlers) handler({ type, payload });
	}
}

function parseEventPayload(raw: string): unknown {
	if (!raw) return {};
	try {
		return JSON.parse(raw);
	} catch {
		return {};
	}
}

export function createEventStream<TEvents extends Record<string, unknown> = XuvaEventPayloadMap>(
	options?: EventStreamOptions<TEvents>
): EventStream<TEvents> {
	return new EventStream<TEvents>(options);
}
