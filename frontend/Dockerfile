FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json ./

# Force install exact version and remove any cached versions
RUN npm cache clean --force && \
    npm uninstall pdfjs-dist 2>/dev/null || true && \
    npm install pdfjs-dist@2.16.105 --save-exact && \
    npm install --legacy-peer-deps

# Verify installed version
RUN npm list pdfjs-dist

# Copy source code
COPY . .

# Build
RUN npm run build

# Runtime image
FROM nginx:alpine

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
