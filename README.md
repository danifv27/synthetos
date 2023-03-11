# synthetos
Synthetos meaning "made up of parts"

docker run -p 8082:8082 -p 8081:8081 --mount type=bind,source=./internal/infrastructure/exporters/features/login.feature,target=/app/features/login.feature,readonly danifv27/uxperi:local test --logging.level=debug --test.features-folder="/app/features"
