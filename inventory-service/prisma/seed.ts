import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

async function main() {
  console.log("Seeding inventory data...");

  const items = [
    { name: "Sony PlayStation 5", quantity: 15 },
    { name: "Xbox Series X", quantity: 10 },
    { name: "Nintendo Switch OLED", quantity: 20 },
    { name: "Apple iPad Air M2", quantity: 8 },
    { name: "Sony WH-1000XM5 Headphones", quantity: 25 },
    { name: "iPhone 15 Pro", quantity: 12 },
  ];

  for (const item of items) {
    const existing = await prisma.item.findFirst({
      where: { name: item.name },
    });

    if (!existing) {
      const created = await prisma.item.create({
        data: item,
      });
      console.log(`Created item: ${created.name} (ID: ${created.id})`);
    } else {
      console.log(`Item already exists: ${existing.name}`);
    }
  }

  console.log("Inventory seeding completed!");
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
