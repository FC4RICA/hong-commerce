import { describe, expect, it, mock } from "bun:test";
import { CreateItem } from "./CreateItem";
import { ValidationError } from "../errors/ValidationError";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";

describe("CreateItem", () => {
  const mockRepo = {
    create: mock(async (data: any) => ({
      id: "1",
      ...data,
      createdAt: new Date(),
      updatedAt: new Date(),
    })),
  } as unknown as ItemRepository;

  const useCase = new CreateItem(mockRepo);

  it("should create an item with valid input", async () => {
    const input = { name: "Test Item", quantity: 10 };
    const item = await useCase.execute(input);

    expect(item.name).toBe("Test Item");
    expect(item.quantity).toBe(10);
    expect(mockRepo.create).toHaveBeenCalled();
  });

  it("should throw ValidationError if name is empty", async () => {
    const input = { name: "  ", quantity: 10 };
    await expect(useCase.execute(input)).rejects.toThrow(ValidationError);
  });

  it("should throw ValidationError if quantity is negative", async () => {
    const input = { name: "Test", quantity: -1 };
    await expect(useCase.execute(input)).rejects.toThrow(ValidationError);
  });
});
