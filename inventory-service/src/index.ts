import { Elysia } from "elysia";
import { swagger } from "@elysiajs/swagger";
import { CreateItem } from "./application/use-cases/CreateItem";
import { DeleteItem } from "./application/use-cases/DeleteItem";
import { GetItem } from "./application/use-cases/GetItem";
import { ListItems } from "./application/use-cases/ListItems";
import { UpdateItem } from "./application/use-cases/UpdateItem";
import { ReserveStock } from "./application/use-cases/ReserveStock";
import { PrismaItemRepository } from "./infrastructure/repositories/PrismaItemRepository";
import { rabbitmq } from "./infrastructure/messaging/rabbitmq";
import { RabbitMQConsumer } from "./infrastructure/messaging/RabbitMQConsumer";
import { RabbitMQEventBus } from "./infrastructure/messaging/RabbitMQEventBus";
import { itemRoutes } from "./presentation/http/item.routes";
import { ValidationError } from "./application/errors/ValidationError";

const itemRepository = new PrismaItemRepository();
const eventBus = new RabbitMQEventBus();

const createItem = new CreateItem(itemRepository);
const listItems = new ListItems(itemRepository);
const getItem = new GetItem(itemRepository);
const updateItem = new UpdateItem(itemRepository);
const deleteItem = new DeleteItem(itemRepository);
const reserveStock = new ReserveStock(itemRepository);

// Start RabbitMQ Consumer
const rabbitMQConsumer = new RabbitMQConsumer(reserveStock, eventBus);
rabbitMQConsumer.start().catch((err) => {
  console.error("[RabbitMQ] Failed to start consumer:", err);
});

const port = process.env.API_PORT ? parseInt(process.env.API_PORT) : 3000;

const app = new Elysia()
  .error({
    VALIDATION_ERROR: ValidationError,
  })
  .onError(({ code, error, set }) => {
    if (code === "VALIDATION_ERROR") {
      set.status = 400;
      return { error: error.message };
    }
  })
  .use(
    swagger({
      documentation: {
        info: {
          title: "Inventory Service",
          version: "1.0.50",
        },
      },
    }),
  )
  .get("/", () => "Inventory service is running.")
  .use(
    itemRoutes({
      createItem,
      listItems,
      getItem,
      updateItem,
      deleteItem,
    }),
  )
  .listen(port);

console.log(
  `\uD83E\uDDBA Elysia is running at ${app.server?.hostname}:${app.server?.port}`,
);

const shutdown = async () => {
  await rabbitmq.close();
  process.exit(0);
};

process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);
