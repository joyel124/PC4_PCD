#La imagen base
FROM golang:alpine
#renombrar el código algoritmo distribuido
COPY ./client.go ./api-svc003.go

#subir archivo csv
# RUN mkdir /var/my-data
# COPY ./customers-100.csv /var/my-data
# RUN chmod 777 /var/my-data/customers-100.csv

#Exponer puerto q usa el algoritmo distribuido
EXPOSE 9002

#ejecutar el algoritmo dentro del contenedor
CMD ["go","run","api-svc003.go"]

