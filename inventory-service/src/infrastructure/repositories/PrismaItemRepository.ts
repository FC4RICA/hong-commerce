import { Prisma } from "@prisma/client";
import type { Item } from "../../domain/entities/Item";
import type {
  CreateItemData,
  ItemRepository,
  UpdateItemData,
} from "../../domain/repositories/ItemRepository";
import { prisma } from "../prisma/client";

export class PrismaItemRepository implements ItemRepository {
  async create(data: CreateItemData): Promise<Item> {
    return prisma.item.create({ data });
  }

  async findById(id: string): Promise<Item | null> {
    return prisma.item.findUnique({ where: { id } });
  }

  async list(): Promise<Item[]> {
    return prisma.item.findMany({ orderBy: { createdAt: "desc" } });
  }

  async update(id: string, data: UpdateItemData): Promise<Item | null> {
    try {
      return await prisma.item.update({ where: { id }, data });
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
      return await prisma.item.delete({ where: { id } });
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
