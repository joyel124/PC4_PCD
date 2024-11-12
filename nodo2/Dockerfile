#La imagen base
FROM golang:alpine
#renombrar el código algoritmo distribuido
COPY ./client.go ./api-svc002.go

#subir archivo csv
RUN mkdir /var/my-data
COPY ./dataset.csv /var/my-data
RUN chmod 777 /var/my-data/dataset.csv

#Exponer puerto q usa el algoritmo distribuido
EXPOSE 9002

#ejecutar el algoritmo dentro del contenedor
CMD ["go","run","api-svc002.go"]

