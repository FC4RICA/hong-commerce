import type { Item } from "../../domain/entities/Item";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";

export class GetItem {
  constructor(private readonly itemRepository: ItemRepository) {}

  async execute(id: string): Promise<Item | null> {
    return this.itemRepository.findById(id);
  }
}
