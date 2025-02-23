FROM postgres:latest

# Set environment variables
#ENV POSTGRES_USER=myuser
#ENV POSTGRES_PASSWORD=password
#ENV POSTGRES_DB=companies

# Copy initialization scripts
COPY ./initdb.d /docker-entrypoint-initdb.d/
