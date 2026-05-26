import { rabbitmq } from "./rabbitmq";
import type { ReserveStock } from "../../application/use-cases/ReserveStock";
import type { EventBus } from "../../application/ports/EventBus";

const ORDERS_EXCHANGE = "order.events";
const QUEUE_NAME = "inventory.order_events";

export class RabbitMQConsumer {
  constructor(
    private readonly reserveStock: ReserveStock,
    private readonly eventBus: EventBus,
  ) {}

  async start(): Promise<void> {
    const channel = await rabbitmq.getChannel();

    await channel.assertExchange(ORDERS_EXCHANGE, "topic", { durable: true });
    await channel.assertQueue(QUEUE_NAME, { durable: true });
    await channel.bindQueue(QUEUE_NAME, ORDERS_EXCHANGE, "order.created");

    console.log(`[RabbitMQ] Listening for events on queue: ${QUEUE_NAME}`);

    channel.consume(QUEUE_NAME, async (msg) => {
      if (!msg) return;

      try {
        const content = JSON.parse(msg.content.toString());
        const routingKey = msg.fields.routingKey;

        console.log(`[RabbitMQ] Received event ${routingKey}:`, content);

        if (routingKey === "order.created") {
          const { id: orderId, items } = content;

          if (!orderId || !Array.isArray(items)) {
            console.error("[RabbitMQ] Invalid order event payload");
            channel.ack(msg);
            return;
          }

          try {
            // 1. Execute Reservation
            await this.reserveStock.execute({
              orderId,
              items: items.map((i: any) => ({
                productId: i.id,
                quantity: i.quantity,
              })),
            });

            // 2. Publish inventory.reserved
            await this.eventBus.publish("inventory.reserved", {
              orderId,
              items,
            });

            console.log(
              `[RabbitMQ] Successfully reserved stock for order ${orderId}`,
            );
            channel.ack(msg);
          } catch (error: any) {
            console.error(
              `[RabbitMQ] Reservation failed for order ${orderId}:`,
              error.message,
            );

            // Publish inventory.failed event so other services can react (Saga compensation)
            await this.eventBus.publish("inventory.failed", {
              orderId,
              reason: error.message,
              timestamp: new Date().toISOString(),
            });

            // Ack the message because we've handled the failure by notifying other services.
            // Nacking without requeueing (false, false) is also an option, but publishing
            // a failure event is the standard way to handle business logic failures in a saga.
            channel.ack(msg);
          }
        } else {
          channel.ack(msg);
        }
      } catch (error) {
        console.error("[RabbitMQ] Error processing message:", error);
        channel.nack(msg, false, false);
      }
    });
  }
}
