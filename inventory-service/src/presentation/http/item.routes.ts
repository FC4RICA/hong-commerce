import { Elysia, t } from "elysia";
import type { CreateItem } from "../../application/use-cases/CreateItem";
import type { DeleteItem } from "../../application/use-cases/DeleteItem";
import type { GetItem } from "../../application/use-cases/GetItem";
import type { ListItems } from "../../application/use-cases/ListItems";
import type { UpdateItem } from "../../application/use-cases/UpdateItem";

export type ItemRoutesDeps = {
  createItem: CreateItem;
  listItems: ListItems;
  getItem: GetItem;
  updateItem: UpdateItem;
  deleteItem: DeleteItem;
};

export const itemRoutes = (deps: ItemRoutesDeps) =>
  new Elysia({ prefix: "/items" })
    .post(
      "/",
      async ({ body, set }) => {
        const item = await deps.createItem.execute(body);
        set.status = 201;
        return item;
      },
      {
        body: t.Object({
          name: t.String(),
          quantity: t.Number(),
        }),
      },
    )
    .get("/", async () => deps.listItems.execute())
    .get("/:id", async ({ params, set }) => {
      const item = await deps.getItem.execute(params.id);
      if (!item) {
        set.status = 404;
        return { error: "Item not found." };
      }
      return item;
    })
    .put(
      "/:id",
      async ({ params, body, set }) => {
        const item = await deps.updateItem.execute({
          id: params.id,
          ...body,
        });

        if (!item) {
          set.status = 404;
          return { error: "Item not found." };
        }

        return item;
      },
      {
        body: t.Object({
          name: t.Optional(t.String()),
          quantity: t.Optional(t.Number()),
        }),
      },
    )
    .delete("/:id", async ({ params, set }) => {
      const item = await deps.deleteItem.execute(params.id);
      if (!item) {
        set.status = 404;
        return { error: "Item not found." };
      }
      return item;
    });
