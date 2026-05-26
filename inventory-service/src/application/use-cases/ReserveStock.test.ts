import { describe, expect, it, mock } from "bun:test";
import { ReserveStock } from "./ReserveStock";
import { ValidationError } from "../errors/ValidationError";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";

// Mocking prisma transaction for testing use case logic
// In a real integration test we would use a test DB, 
// but for unit test we mock the internal prisma client call used in the use case.
import { prisma } from "../../infrastructure/prisma/client";

mock.module("../../infrastructure/prisma/client", () => ({
  prisma: {
    $transaction: mock(async (callback: any) => {
      const tx = {
        item: {
          findUnique: mock(async ({ where }: any) => {
            if (where.id === "1") return { id: "1", name: "Item 1", quantity: 10, reserved: 2 };
            return null;
          }),
          update: mock(async ({ where, data }: any) => ({
             id: where.id,
             name: "Item 1",
             quantity: 10,
             reserved: 2 + data.reserved.increment,
             createdAt: new Date(),
             updatedAt: new Date(),
          })),
        },
        reservation: {
          create: mock(async () => ({})),
        },
      };
      return callback(tx);
    }),
  },
}));

describe("ReserveStock", () => {
  const mockRepo = {} as unknown as ItemRepository;
  const useCase = new ReserveStock(mockRepo);

  it("should reserve stock successfully when available", async () => {
    const input = {
      orderId: "order-123",
      items: [{ productId: "1", quantity: 3 }],
    };

    const result = await useCase.execute(input);

    expect(result.length).toBe(1);
    expect(result[0].reserved).toBe(5);
    expect(result[0].available).toBe(5); // 10 - 5
  });

  it("should throw ValidationError when insufficient stock", async () => {
    const input = {
      orderId: "order-123",
      items: [{ productId: "1", quantity: 9 }], // 10 - 2 = 8 available
    };

    await expect(useCase.execute(input)).rejects.toThrow(/Insufficient stock/);
  });

  it("should throw ValidationError when item not found", async () => {
    const input = {
      orderId: "order-123",
      items: [{ productId: "99", quantity: 1 }],
    };

    await expect(useCase.execute(input)).rejects.toThrow(/Item 99 not found/);
  });
});
