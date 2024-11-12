# Collaborative Filtering in Recommender System

Este proyecto implementa una red distribuida donde el nodo servidor se encarga de un usuario, mientras que los nodos cliente se encargan de procesar los datos y devolver las recomendaciones para ese usuario al nodo servidor.

## Estructura del Proyecto

- **`nodo1|nodo2|nodo3`**: Carpetas que contiene la implementación del nodo cliente con su respectivo Dockerfile.
- **`server`**: Carpeta que contiene la implementación del nodo servidor con su respectivo Dockerfile.
- **`nodo1|nodo2|nodo3/dataset.csv`**: Dataset de valoracion de peliculas(UserID: Id del usuario; MovieID: Id de la pelicula; Rating: Valoracion de la pelicula hecha por el usuario).
- **`docker-compose.yml`**: Archivo con la configuracion de los contenedores(nodo1, nodo2, nodo3 y server).
- **`test.go`**: Archivo de prueba que contiene la implementacion del filtro colaborativo.

## Requisitos

- Go (Golang) instalado en tu sistema.
- El archivo de datos `dataset.csv` dentro de la carpeta nodo1/nodo2/nodo3(se tiene que descargar por separado ya que no se podia subir al repositorio porque excedia el limite).
- Docker.

## Instalación

1. Clona el repositorio:
    ```bash
    git clone https://github.com/joyel124/PC4_PCD.git
    cd PC4_PCD
    ```
2. Descargar el dataset que se encuentra en el siguiente enlace: https://drive.google.com/drive/folders/1dkMcvRyeZWavG3uMAS0iw9lwosWsAEH7?usp=sharing.
3. Copia el archivo `dataset.csv` y ponlo dentro de la carpeta nodo1/nodo2/nodo3 del proyecto.
   
## Ejecución

Puedes ejecutar el programa principal usando el comando:

```bash
docker-compose up --build
