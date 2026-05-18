import type { Item } from "../../domain/entities/Item";
import type {
  ItemRepository,
  UpdateItemData,
} from "../../domain/repositories/ItemRepository";
import { ValidationError } from "../errors/ValidationError";

export type UpdateItemInput = {
  id: string;
  name?: string;
  quantity?: number;
};

export class UpdateItem {
  constructor(private readonly itemRepository: ItemRepository) {}

  async execute(input: UpdateItemInput): Promise<Item | null> {
    const data: UpdateItemData = {};

    if (input.name !== undefined) {
      const name = input.name.trim();
      if (!name) {
        throw new ValidationError("Name cannot be empty.");
      }
      data.name = name;
    }

    if (input.quantity !== undefined) {
      if (!Number.isInteger(input.quantity) || input.quantity < 0) {
        throw new ValidationError("Quantity must be a non-negative integer.");
      }
      data.quantity = input.quantity;
    }

    if (Object.keys(data).length === 0) {
      throw new ValidationError(
        "At least one field must be provided for update.",
      );
    }

    return this.itemRepository.update(input.id, data);
  }
}
