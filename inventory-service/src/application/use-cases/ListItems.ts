import type { Item } from "../../domain/entities/Item";
import type { ItemRepository } from "../../domain/repositories/ItemRepository";

export class ListItems {
  constructor(private readonly itemRepository: ItemRepository) {}

  async execute(): Promise<Item[]> {
    return this.itemRepository.list();
  }
}
