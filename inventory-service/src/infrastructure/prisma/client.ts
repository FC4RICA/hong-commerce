import { PrismaClient } from "@prisma/client";

const dbUser = process.env.DB_USER || "postgres";
const dbPassword = process.env.DB_PASSWORD || "postgres";
const dbHost = process.env.DB_HOST || "localhost";
const dbPort = process.env.DB_PORT || "5432";
const dbName = process.env.DB_NAME || "inventorydb";

const url = `postgresql://${dbUser}:${dbPassword}@${dbHost}:${dbPort}/${dbName}?sslmode=disable`;

export const prisma = new PrismaClient({
  datasources: {
    db: {
      url,
    },
  },
});
