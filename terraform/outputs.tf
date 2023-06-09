output "endpoint" {
  value       = google_cloud_run_service.fe.status[0].url
  description = "URL del frontend de la aplicación"
}

output "sqlservername" {
  value       = google_sql_database_instance.main.name
  description = "Nombre del servidor de la base de datos."
}

output "api" {
  value       = google_cloud_run_service.api.status[0].url
  description = "URL de la API de la aplicación"
}
