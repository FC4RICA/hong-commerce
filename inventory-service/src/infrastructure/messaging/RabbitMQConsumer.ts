import { rabbitmq } from "./rabbitmq";
import type { DecreaseStock } from "../../application/use-cases/DecreaseStock";

const ORDERS_EXCHANGE = "order.events";
const QUEUE_NAME = "inventory.order_events";

export class RabbitMQConsumer {
  constructor(private readonly decreaseStock: DecreaseStock) {}

  async start(): Promise<void> {
    const channel = await rabbitmq.getChannel();

    // Ensure the exchange for orders exists
    await channel.assertExchange(ORDERS_EXCHANGE, "topic", { durable: true });

    // Create a queue for inventory service to listen to order events
    await channel.assertQueue(QUEUE_NAME, { durable: true });

    // Bind the queue to the exchange with a specific routing key
    await channel.bindQueue(QUEUE_NAME, ORDERS_EXCHANGE, "order.created");

    console.log(`[RabbitMQ] Listening for events on queue: ${QUEUE_NAME}`);

    channel.consume(QUEUE_NAME, async (msg) => {
      if (!msg) return;

      try {
        const content = JSON.parse(msg.content.toString());
        const routingKey = msg.fields.routingKey;

        console.log(`[RabbitMQ] Received event ${routingKey}:`, content);

        if (routingKey === "order.created") {
          // Assume the order event payload has items: [{ id: string, quantity: number }]
          const { items } = content;
          if (Array.isArray(items)) {
            for (const orderItem of items) {
              await this.decreaseStock.execute({
                itemId: orderItem.id,
                quantity: orderItem.quantity,
              });
            }
          }
        }

        channel.ack(msg);
      } catch (error) {
        console.error("[RabbitMQ] Error processing message:", error);
        // In a real app, you might want to nack with requeue: false and send to a dead-letter-queue
        channel.nack(msg, false, false);
      }
    });
  }
}
