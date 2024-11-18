#La imagen base
FROM golang:alpine
#renombrar el código algoritmo distribuido
COPY ./server.go ./api-svc004.go

# subir archivo csv
# RUN mkdir /var/my-data
# COPY ./dataset.csv /var/my-data
# RUN chmod 777 /var/my-data/dataset.csv

# subir 3 archivos csv
RUN mkdir /var/my-data
COPY ./dataset_1.csv /var/my-data
COPY ./dataset_2.csv /var/my-data
COPY ./dataset_3.csv /var/my-data
RUN chmod 777 /var/my-data/dataset_1.csv
RUN chmod 777 /var/my-data/dataset_2.csv
RUN chmod 777 /var/my-data/dataset_3.csv

#Exponer puerto q usa el algoritmo distribuido
EXPOSE 9002

#ejecutar el algoritmo dentro del contenedor
CMD ["go","run","api-svc004.go"]

