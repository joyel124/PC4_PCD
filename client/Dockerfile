# Etapa 1: Construcción de la aplicación
FROM node:18 AS build

WORKDIR /app

# Copiar los archivos de dependencias
COPY my-app/package*.json ./
RUN npm install

# Copiar todo el código fuente de my-app
COPY my-app/ ./

# Construir la aplicación
RUN npm run build

# Etapa 2: Servir la aplicación en producción
FROM node:18

WORKDIR /app

# Copiar los archivos de la etapa de construcción
COPY --from=build /app/ ./

# Exponer el puerto 3000 para Next.js
EXPOSE 3000

# Iniciar el servidor Next.js en producción
CMD ["npm", "run", "start"]
