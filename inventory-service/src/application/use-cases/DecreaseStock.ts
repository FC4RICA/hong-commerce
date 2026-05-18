import type { Item } from "../../domain/entities/Item";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";
import { ValidationError } from "../errors/ValidationError";

export type DecreaseStockInput = {
  itemId: string;
  quantity: number;
};

export class DecreaseStock {
  constructor(private readonly itemRepository: ItemRepository) {}

  async execute(input: DecreaseStockInput): Promise<Item | null> {
    const item = await this.itemRepository.findById(input.itemId);
    if (!item) {
      console.warn(`[DecreaseStock] Item ${input.itemId} not found.`);
      return null;
    }

    if (item.quantity < input.quantity) {
      throw new ValidationError(`Insufficient stock for item ${item.name}. Available: ${item.quantity}, Requested: ${input.quantity}`);
    }

    const updatedItem = await this.itemRepository.update(input.itemId, {
      quantity: item.quantity - input.quantity,
    });

    console.log(`[DecreaseStock] Updated item ${item.name} stock. New quantity: ${updatedItem?.quantity}`);

    return updatedItem;
  }
}
