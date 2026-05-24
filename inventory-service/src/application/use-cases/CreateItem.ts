import type { Item } from "../../domain/entities/Item";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";
import { ValidationError } from "../errors/ValidationError";

export type CreateItemInput = {
  name: string;
  quantity: number;
  reserved?: number;
};

export class CreateItem {
  constructor(private readonly itemRepository: ItemRepository) {}

  async execute(input: CreateItemInput): Promise<Item> {
    const name = input.name.trim();
    if (!name) {
      throw new ValidationError("Name is required.");
    }

    if (!Number.isInteger(input.quantity) || input.quantity < 0) {
      throw new ValidationError("Quantity must be a non-negative integer.");
    }

    if (input.reserved !== undefined && (!Number.isInteger(input.reserved) || input.reserved < 0)) {
        throw new ValidationError("Reserved must be a non-negative integer.");
    }

    return this.itemRepository.create({
      name,
      quantity: input.quantity,
      reserved: input.reserved ?? 0,
    });
  }
}
