import { describe, expect, it, mock } from "bun:test";
import { DecreaseStock } from "./DecreaseStock";
import { ValidationError } from "../errors/ValidationError";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";

describe("DecreaseStock", () => {
  const mockRepo = {
    findById: mock(async (id: string) => {
      if (id === "1") return { id: "1", name: "Test Item", quantity: 10 };
      return null;
    }),
    update: mock(async (id: string, data: any) => ({
      id,
      name: "Test Item",
      ...data,
    })),
  } as unknown as ItemRepository;

  const useCase = new DecreaseStock(mockRepo);

  it("should decrease stock correctly", async () => {
    const input = { itemId: "1", quantity: 3 };
    const result = await useCase.execute(input);

    expect(result?.quantity).toBe(7);
    expect(mockRepo.update).toHaveBeenCalledWith("1", { quantity: 7 });
  });

  it("should throw ValidationError if insufficient stock", async () => {
    const input = { itemId: "1", quantity: 15 };
    await expect(useCase.execute(input)).rejects.toThrow(ValidationError);
    await expect(useCase.execute(input)).rejects.toThrow(/Insufficient stock/);
  });

  it("should return null if item not found", async () => {
    const input = { itemId: "99", quantity: 1 };
    const result = await useCase.execute(input);
    expect(result).toBeNull();
  });
});
