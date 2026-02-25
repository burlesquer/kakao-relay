#!/bin/bash
set -e

# Configuration - All values MUST be set via environment variables
SERVICE_NAME="${SERVICE_NAME:-kakao-talkchannel-relay}"
REGION="${REGION:?REGION is required}"
PROJECT_ID="${PROJECT_ID:?PROJECT_ID is required}"
CLOUD_SQL_INSTANCE="${CLOUD_SQL_INSTANCE:?CLOUD_SQL_INSTANCE is required}"
VPC_CONNECTOR="${VPC_CONNECTOR:?VPC_CONNECTOR is required}"
IMAGE_NAME="${REGION}-docker.pkg.dev/${PROJECT_ID}/cloud-run-source-deploy/${SERVICE_NAME}"

ENV_VARS="LOG_LEVEL=info,CALLBACK_TTL_SECONDS=55"
if [[ -n "${PORTAL_BASE_URL:-}" ]]; then
  ENV_VARS="${ENV_VARS},PORTAL_BASE_URL=${PORTAL_BASE_URL}"
fi

echo "Deploying $SERVICE_NAME to Cloud Run..."
echo "  Project: $PROJECT_ID"
echo "  Region:  $REGION"

# Build frontend assets
echo ""
echo "Building frontend assets..."
bun run build:admin
bun run build:portal
echo "Frontend build complete."
echo ""

# Step 1: Build and push image (no cache issues)
echo "Building Docker image..."
gcloud builds submit --tag "${IMAGE_NAME}:latest" --project "$PROJECT_ID" .
echo "Image build complete."
echo ""

# Step 2: Deploy to Cloud Run
echo "Deploying to Cloud Run..."
gcloud run deploy "$SERVICE_NAME" \
  --image "${IMAGE_NAME}:latest" \
  --project "$PROJECT_ID" \
  --region "$REGION" \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 3 \
  --concurrency 100 \
  --timeout 60s \
  --cpu-boost \
  --add-cloudsql-instances "$CLOUD_SQL_INSTANCE" \
  --vpc-connector "$VPC_CONNECTOR" \
  --set-env-vars "${ENV_VARS}" \
  --set-secrets "DATABASE_URL=${SECRET_PREFIX:-kakao-relay}-database-url:latest,\
REDIS_URL=${SECRET_PREFIX:-kakao-relay}-redis-url:latest,\
ADMIN_PASSWORD=${SECRET_PREFIX:-kakao-relay}-admin-password:latest,\
PORTAL_SESSION_SECRET=${SECRET_PREFIX:-kakao-relay}-session-secret:latest,\
ADMIN_SESSION_SECRET=${SECRET_PREFIX:-kakao-relay}-admin-session-secret:latest"

echo ""
echo "Deployment complete!"
gcloud run services describe "$SERVICE_NAME" --region "$REGION" --format 'value(status.url)'
