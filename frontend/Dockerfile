# Dev-mode Vite server — for a production build swap the last stage
# for an nginx serving `npm run build` output.
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
