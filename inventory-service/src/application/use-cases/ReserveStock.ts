import type { ItemRepository } from "../../domain/repositories/ItemRepository";
import { ValidationError } from "../errors/ValidationError";
import { prisma } from "../../infrastructure/prisma/client";
import { mapToItem } from "../../domain/entities/Item";

export type ReserveStockInput = {
  orderId: string;
  items: Array<{
    productId: string;
    quantity: number;
  }>;
};

export class ReserveStock {
  constructor(private readonly itemRepository: ItemRepository) {}

  async execute(input: ReserveStockInput) {
    return await prisma.$transaction(async (tx) => {
      const reservedItems = [];

      for (const orderItem of input.items) {
        const item = await tx.item.findUnique({
          where: { id: orderItem.productId },
        });

        if (!item) {
          throw new ValidationError(`Item ${orderItem.productId} not found.`);
        }

        const available = item.quantity - item.reserved;
        if (available < orderItem.quantity) {
          throw new ValidationError(
            `Insufficient stock for item ${item.name}. Available: ${available}, Requested: ${orderItem.quantity}`
          );
        }

        // 1. Create Reservation record
        await tx.reservation.create({
          data: {
            orderId: input.orderId,
            itemId: item.id,
            quantity: orderItem.quantity,
            status: "reserved",
          },
        });

        // 2. Increment reserved count on Item
        const updatedItem = await tx.item.update({
          where: { id: item.id },
          data: {
            reserved: {
              increment: orderItem.quantity,
            },
          },
        });

        reservedItems.push(mapToItem(updatedItem));
      }

      return reservedItems;
    });
  }
}
