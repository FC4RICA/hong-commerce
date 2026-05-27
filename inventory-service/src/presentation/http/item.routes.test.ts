import { describe, expect, it, mock } from "bun:test";
import { Elysia } from "elysia";
import { itemRoutes } from "./item.routes";
import { ValidationError } from "../../application/errors/ValidationError";

describe("Item Routes", () => {
  const mockDeps = {
    createItem: {
      execute: mock(async (body: any) => ({ id: "1", ...body })),
    },
    listItems: {
      execute: mock(async () => [{ id: "1", name: "Test", quantity: 10 }]),
    },
    getItem: {
      execute: mock(async (id: string) => {
        if (id === "1") return { id: "1", name: "Test", quantity: 10 };
        return null;
      }),
    },
    updateItem: {
      execute: mock(async ({ id, ...body }: any) => {
        if (id === "1") return { id: "1", ...body };
        return null;
      }),
    },
    deleteItem: {
      execute: mock(async (id: string) => {
        if (id === "1") return { id: "1", name: "Test", quantity: 10 };
        return null;
      }),
    },
  } as any;

  const app = new Elysia()
    .error({ VALIDATION_ERROR: ValidationError })
    .onError(({ code, error, set }) => {
      if (code === "VALIDATION_ERROR") {
        set.status = 400;
        return { error: error.message };
      }
    })
    .use(itemRoutes(mockDeps));

  it("POST /items should create an item", async () => {
    const response = await app.handle(
      new Request("http://localhost/items", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "New Item", quantity: 5 }),
      }),
    );

    expect(response.status).toBe(201);
    const data = await response.json();
    expect(data.name).toBe("New Item");
  });

  it("POST /items should return 400 on validation error", async () => {
    mockDeps.createItem.execute.mockImplementationOnce(() => {
      throw new ValidationError("Invalid name");
    });

    const response = await app.handle(
      new Request("http://localhost/items", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "", quantity: 5 }),
      }),
    );

    expect(response.status).toBe(400);
    const data = await response.json();
    expect(data.error).toBe("Invalid name");
  });

  it("GET /items should list items", async () => {
    const response = await app.handle(new Request("http://localhost/items"));
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(Array.isArray(data)).toBe(true);
    expect(data.length).toBe(1);
  });

  it("GET /items/:id should return an item", async () => {
    const response = await app.handle(new Request("http://localhost/items/1"));
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.id).toBe("1");
  });

  it("GET /items/:id should return 404 if not found", async () => {
    const response = await app.handle(new Request("http://localhost/items/99"));
    expect(response.status).toBe(404);
  });

  it("PUT /items/:id should update an item", async () => {
    const response = await app.handle(
      new Request("http://localhost/items/1", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "Updated" }),
      }),
    );

    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.name).toBe("Updated");
  });

  it("DELETE /items/:id should delete an item", async () => {
    const response = await app.handle(
      new Request("http://localhost/items/1", {
        method: "DELETE",
      }),
    );

    expect(response.status).toBe(200);
  });
});
