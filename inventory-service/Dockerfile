# 1. Base image for shared configuration
FROM oven/bun:1.1-slim AS base
WORKDIR /app

# 2. Dependency installation stage
FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock /temp/dev/
RUN cd /temp/dev && bun install --frozen-lockfile

# Install production dependencies only
RUN mkdir -p /temp/prod
COPY package.json bun.lock /temp/prod/
RUN cd /temp/prod && bun install --frozen-lockfile --production

# 3. Development stage
FROM base AS dev
COPY --from=install /temp/dev/node_modules node_modules
COPY . .
ENV NODE_ENV=development
CMD ["bun", "run", "dev"]

# 4. Prerelease/Build stage (generate Prisma)
FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules
COPY . .
RUN bun run db:generate

# 5. Production/Release stage
FROM base AS prod
ENV NODE_ENV=production
# Copy production node_modules and generated prisma client
COPY --from=install /temp/prod/node_modules node_modules
COPY --from=prerelease /app/node_modules/@prisma/client node_modules/@prisma/client
COPY --from=prerelease /app/src src
COPY --from=prerelease /app/package.json .
COPY --from=prerelease /app/prisma prisma

# Run as non-root user
USER bun
EXPOSE 3000
CMD ["bun", "run", "src/index.ts"]
