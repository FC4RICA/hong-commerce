import { Prisma } from "@prisma/client";
import { type Item, mapToItem } from "../../domain/entities/Item";
import type {
  CreateItemData,
  ItemRepository,
  UpdateItemData,
} from "../../domain/repositories/ItemRepository";
import { prisma } from "../prisma/client";

export class PrismaItemRepository implements ItemRepository {
  async create(data: CreateItemData): Promise<Item> {
    const item = await prisma.item.create({ data });
    return mapToItem(item);
  }

  async findById(id: string): Promise<Item | null> {
    const item = await prisma.item.findUnique({ where: { id } });
    return item ? mapToItem(item) : null;
  }

  async list(): Promise<Item[]> {
    const items = await prisma.item.findMany({ orderBy: { createdAt: "desc" } });
    return items.map(mapToItem);
  }

  async update(id: string, data: UpdateItemData): Promise<Item | null> {
    try {
      const item = await prisma.item.update({ where: { id }, data });
      return mapToItem(item);
    } catch (error) {
      if (
        error instanceof Prisma.PrismaClientKnownRequestError &&
        error.code === "P2025"
      ) {
        return null;
      }
      throw error;
    }
  }

  async delete(id: string): Promise<Item | null> {
    try {
      const item = await prisma.item.delete({ where: { id } });
      return item ? mapToItem(item) : null;
    } catch (error) {
      if (
        error instanceof Prisma.PrismaClientKnownRequestError &&
        error.code === "P2025"
      ) {
        return null;
      }
      throw error;
    }
  }
}
