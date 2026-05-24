import type { EventBus, EventPayload } from "../../application/ports/EventBus";
import { rabbitmq } from "./rabbitmq";

export class RabbitMQEventBus implements EventBus {
  async publish(event: string, payload: EventPayload): Promise<void> {
    const channel = await rabbitmq.getChannel();
    const message = Buffer.from(JSON.stringify(payload));
    channel.publish(rabbitmq.exchange, event, message, {
      contentType: "application/json",
      persistent: true,
      timestamp: Date.now(),
      type: event,
    });
  }
}
