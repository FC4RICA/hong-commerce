export type EventPayload = Record<string, unknown>;

export interface EventBus {
  publish(event: string, payload: EventPayload): Promise<void>;
}
