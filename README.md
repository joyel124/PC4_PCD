# Collaborative Filtering in Recommender System

Este proyecto implementa una red distribuida donde el nodo servidor se encarga de leer y fragmetar los datos, mientras que los nodos cliente se encargan de procesar los datos y devolver las recomendaciones al nodo servidor.

## Estructura del Proyecto

- **`nodo1|nodo2|nodo3`**: Carpetas que contiene la implementación del nodo cliente con su respectivo Dockerfile.
- **`server`**: Carpeta que contiene la implementación del nodo servidor con su respectivo Dockerfile.
- **`server/movies_data.csv`**: Dataset valoracion de peliculas(UserID: Id del usuario; MovieID: Id de la pelicula; Rating: Valoracion de la pelicula hecha por el usuario).
- **`docker-compose.yml`**: Archivo con la configuracion de los contenedores(nodo1, nodo2, nodo3 y server).

## Requisitos

- Go (Golang) instalado en tu sistema.
- El archivo de datos `movies_data.csv` dentro de la carpeta server(se tiene que descargar por separado ya que no se podia subir al repositorio porque excedia el limite).
- Docker.

## Instalación

1. Clona el repositorio:
    ```bash
    git clone https://github.com/joyel124/PC4_PCD.git
    cd PC4_PCD
    ```

2. Asegúrate de que tienes el archivo `movies_data.csv` dentro de la carpeta server.
   
## Ejecución

Puedes ejecutar el programa principal usando el comando:

```bash
docker-compose up --build
