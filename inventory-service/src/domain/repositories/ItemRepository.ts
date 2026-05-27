import type { Item } from "../entities/Item";

export type CreateItemData = {
  name: string;
  quantity: number;
  reserved?: number;
};

export type UpdateItemData = {
  name?: string;
  quantity?: number;
  reserved?: number;
};

export interface ItemRepository {
  create(data: CreateItemData): Promise<Item>;
  findById(id: string): Promise<Item | null>;
  list(): Promise<Item[]>;
  update(id: string, data: UpdateItemData): Promise<Item | null>;
  delete(id: string): Promise<Item | null>;
}
