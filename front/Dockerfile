# Multi-stage build for frontend
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /build

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Accept build arg for API URL (empty string for relative URLs)
ARG VITE_API_URL=
ENV VITE_API_URL=${VITE_API_URL}

# Build the application
RUN npm run build

# Production stage with Nginx
FROM nginx:1.27-alpine

# Copy built assets from builder
COPY --from=builder /build/dist /usr/share/nginx/html

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
