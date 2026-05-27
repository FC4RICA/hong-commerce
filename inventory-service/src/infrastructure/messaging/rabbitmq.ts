import amqplib, { type Channel, type Connection } from "amqplib";

let connection: Connection | null = null;
let channel: Channel | null = null;

const EXCHANGE = "inventory.events";

export const rabbitmq = {
  exchange: EXCHANGE,
  async getChannel(): Promise<Channel> {
    if (channel) {
      return channel;
    }

    const url = process.env.RABBITMQ_URL ?? "amqp://localhost:5672";
    try {
      connection = await amqplib.connect(url);

      connection.on("error", (err) => {
        console.error("[RabbitMQ] connection error", err);
        connection = null;
        channel = null;
      });

      connection.on("close", () => {
        console.warn("[RabbitMQ] connection closed");
        connection = null;
        channel = null;
      });

      channel = await connection.createChannel();
      await channel.assertExchange(EXCHANGE, "topic", { durable: true });

      return channel;
    } catch (error) {
      console.error("[RabbitMQ] failed to connect", error);
      throw error;
    }
  },
  async close(): Promise<void> {
    if (channel) {
      await channel.close();
      channel = null;
    }
    if (connection) {
      await connection.close();
      connection = null;
    }
  },
};
