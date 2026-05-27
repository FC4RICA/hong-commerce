import { describe, expect, it, mock } from "bun:test";
import { UpdateItem } from "./UpdateItem";
import { ValidationError } from "../errors/ValidationError";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";

describe("UpdateItem", () => {
  const mockRepo = {
    update: mock(async (id: string, data: any) => {
      if (id === "non-existent") return null;
      return {
        id,
        name: data.name ?? "Original Name",
        quantity: data.quantity ?? 10,
        createdAt: new Date(),
        updatedAt: new Date(),
      };
    }),
  } as unknown as ItemRepository;

  const useCase = new UpdateItem(mockRepo);

  it("should update an item with valid input", async () => {
    const input = { id: "1", name: "Updated Name", quantity: 20 };
    const item = await useCase.execute(input);

    expect(item?.name).toBe("Updated Name");
    expect(item?.quantity).toBe(20);
    expect(mockRepo.update).toHaveBeenCalled();
  });

  it("should return null if item does not exist", async () => {
    const input = { id: "non-existent", name: "Updated Name" };
    const item = await useCase.execute(input);
    expect(item).toBeNull();
  });

  it("should throw ValidationError if no fields provided", async () => {
    const input = { id: "1" };
    await expect(useCase.execute(input)).rejects.toThrow("At least one field must be provided for update.");
  });
});
