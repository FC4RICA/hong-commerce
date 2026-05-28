# 1. Base image
FROM oven/bun:latest AS base
WORKDIR /app

# 2. Dependency installation stage
FROM base AS install
COPY package.json bun.lock ./
RUN bun install

# 3. Development stage
FROM base AS dev
COPY --from=install /app/node_modules node_modules
COPY . .
RUN ./node_modules/.bin/prisma generate
ENV NODE_ENV=development
CMD ["sh", "-c", "./node_modules/.bin/prisma db push && bun run dev"]

# 4. Build stage (generate Prisma)
FROM base AS build
COPY --from=install /app/node_modules node_modules
COPY . .
RUN ./node_modules/.bin/prisma generate

# 5. Production stage
FROM base AS prod
ENV NODE_ENV=production
COPY --from=install /app/node_modules node_modules
COPY --from=build /app/src src
COPY --from=build /app/package.json .
COPY --from=build /app/prisma prisma

# Run as non-root user
USER bun
EXPOSE 3000
CMD ["sh", "-c", "./node_modules/.bin/prisma db push && bun run src/index.ts"]
