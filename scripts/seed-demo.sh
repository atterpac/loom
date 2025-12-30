#!/bin/bash

# Tempo Demo Seed Script
# Creates synthetic workflows for demonstrating Tempo TUI features
#
# Prerequisites:
#   - Temporal server running (default: localhost:7233)
#   - Temporal CLI installed (`temporal`)
#   - Demo worker running: go run ./cmd/demo-worker
#
# Usage:
#   1. Start the demo worker: go run ./cmd/demo-worker
#   2. Run this script: ./scripts/seed-demo.sh [--address localhost:7233] [--namespace default]
#
# The demo worker will process workflows with various outcomes:
#   - Successful completions with child workflows
#   - Failures (payment declined, verification failed, etc.)
#   - Multi-step workflows with activities
#   - Long-running workflows with heartbeats

set -e

# Default configuration
ADDRESS="localhost:7233"
NAMESPACE="default"
TASK_QUEUE="demo-queue"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --address)
            ADDRESS="$2"
            shift 2
            ;;
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --task-queue)
            TASK_QUEUE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--address HOST:PORT] [--namespace NAME] [--task-queue NAME]"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}                    Tempo Demo Seed Script                                ${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Address:    ${GREEN}$ADDRESS${NC}"
echo -e "  Namespace:  ${GREEN}$NAMESPACE${NC}"
echo -e "  Task Queue: ${GREEN}$TASK_QUEUE${NC}"
echo ""

# Common temporal CLI args
TEMPORAL_ARGS="--address $ADDRESS --namespace $NAMESPACE"

# Helper function to start a workflow
start_workflow() {
    local workflow_type=$1
    local workflow_id=$2
    local input=$3

    echo -e "  ${YELLOW}→${NC} Starting ${GREEN}$workflow_type${NC} (${workflow_id})"

    if [ -n "$input" ]; then
        temporal workflow start $TEMPORAL_ARGS \
            --type "$workflow_type" \
            --task-queue "$TASK_QUEUE" \
            --workflow-id "$workflow_id" \
            --input "$input" \
            2>/dev/null || echo -e "    ${RED}Failed (may already exist)${NC}"
    else
        temporal workflow start $TEMPORAL_ARGS \
            --type "$workflow_type" \
            --task-queue "$TASK_QUEUE" \
            --workflow-id "$workflow_id" \
            2>/dev/null || echo -e "    ${RED}Failed (may already exist)${NC}"
    fi
}

# Helper to generate random ID suffix
random_suffix() {
    echo $(head -c 4 /dev/urandom | xxd -p)
}

echo -e "${BLUE}[1/6]${NC} Creating Order Processing Workflows..."
echo ""

# E-commerce order workflows
start_workflow "OrderWorkflow" "order-$(random_suffix)" '{"orderId": "ORD-10042", "customerId": "cust_8a7f3b", "items": [{"sku": "WIDGET-001", "qty": 2}, {"sku": "GADGET-PRO", "qty": 1}], "total": 149.99}'
start_workflow "OrderWorkflow" "order-$(random_suffix)" '{"orderId": "ORD-10043", "customerId": "cust_2c9d1e", "items": [{"sku": "SENSOR-X", "qty": 5}], "total": 89.95}'
start_workflow "OrderWorkflow" "order-$(random_suffix)" '{"orderId": "ORD-10044", "customerId": "cust_f4e8a2", "items": [{"sku": "CONTROLLER-V2", "qty": 1}], "total": 299.00}'

echo ""
echo -e "${BLUE}[2/6]${NC} Creating User Management Workflows..."
echo ""

# User registration and management
start_workflow "UserRegistration" "user-reg-$(random_suffix)" '{"email": "alice@example.com", "plan": "pro", "referralCode": "FRIEND20"}'
start_workflow "UserRegistration" "user-reg-$(random_suffix)" '{"email": "bob@example.com", "plan": "starter"}'
start_workflow "UserOnboarding" "onboard-$(random_suffix)" '{"userId": "usr_a1b2c3", "steps": ["welcome", "profile", "preferences", "tutorial"]}'
start_workflow "UserOnboarding" "onboard-$(random_suffix)" '{"userId": "usr_d4e5f6", "steps": ["welcome", "profile", "preferences"]}'
start_workflow "AccountVerification" "verify-$(random_suffix)" '{"userId": "usr_g7h8i9", "method": "email"}'

echo ""
echo -e "${BLUE}[3/6]${NC} Creating Payment Processing Workflows..."
echo ""

# Payment workflows
start_workflow "PaymentProcess" "payment-$(random_suffix)" '{"paymentId": "pay_xyz123", "amount": 99.99, "currency": "USD", "method": "card"}'
start_workflow "PaymentProcess" "payment-$(random_suffix)" '{"paymentId": "pay_abc456", "amount": 250.00, "currency": "USD", "method": "ach"}'
start_workflow "RefundProcess" "refund-$(random_suffix)" '{"refundId": "ref_001", "originalPayment": "pay_old789", "amount": 49.99, "reason": "customer_request"}'
start_workflow "SubscriptionBilling" "billing-$(random_suffix)" '{"subscriptionId": "sub_monthly_001", "customerId": "cust_premium", "amount": 29.99}'

echo ""
echo -e "${BLUE}[4/6]${NC} Creating Data Processing Workflows..."
echo ""

# Data processing / ETL workflows
start_workflow "DataImport" "import-$(random_suffix)" '{"source": "s3://data-lake/exports/2024-01", "format": "parquet", "rowCount": 150000}'
start_workflow "DataExport" "export-$(random_suffix)" '{"destination": "analytics-warehouse", "tables": ["users", "events", "transactions"]}'
start_workflow "ReportGeneration" "report-$(random_suffix)" '{"reportType": "monthly-summary", "period": "2024-01", "format": "pdf"}'
start_workflow "ETLPipeline" "etl-$(random_suffix)" '{"pipeline": "customer-360", "sources": ["crm", "support", "billing"]}'
start_workflow "DataValidation" "validate-$(random_suffix)" '{"dataset": "product-catalog", "rules": ["schema", "completeness", "uniqueness"]}'

echo ""
echo -e "${BLUE}[5/6]${NC} Creating Notification Workflows..."
echo ""

# Notification workflows
start_workflow "EmailCampaign" "campaign-$(random_suffix)" '{"campaignId": "welcome-series", "segment": "new-users", "templateId": "tmpl_welcome_01"}'
start_workflow "NotificationBatch" "notify-$(random_suffix)" '{"type": "push", "audience": "active-users", "message": "New features available!"}'
start_workflow "SMSNotification" "sms-$(random_suffix)" '{"phone": "+1555000####", "template": "order-shipped", "orderId": "ORD-10040"}'

echo ""
echo -e "${BLUE}[6/6]${NC} Creating Long-Running & Scheduled Workflows..."
echo ""

# Long-running workflows (good for showing "Running" state)
start_workflow "LongRunningProcess" "long-process-$(random_suffix)" '{"duration": "24h", "checkpointInterval": "1h"}'
start_workflow "MonitoringWorkflow" "monitor-$(random_suffix)" '{"target": "api-cluster", "interval": "5m", "alertThreshold": 0.95}'
start_workflow "BatchProcessor" "batch-$(random_suffix)" '{"batchId": "batch-2024-001", "itemCount": 10000, "parallelism": 10}'

# Saga / multi-step workflows
start_workflow "OrderSaga" "saga-$(random_suffix)" '{"orderId": "ORD-SAGA-001", "steps": ["reserve-inventory", "charge-payment", "ship-order", "send-confirmation"]}'
start_workflow "ProvisioningWorkflow" "provision-$(random_suffix)" '{"resourceType": "kubernetes-cluster", "config": {"nodes": 3, "region": "us-west-2"}}'

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Done!${NC} Created demo workflows in namespace '${NAMESPACE}'"
echo ""
echo -e "To view in Tempo:"
echo -e "  ${YELLOW}tempo --namespace $NAMESPACE${NC}"
echo ""
echo -e "If the demo worker is running (go run ./cmd/demo-worker), workflows will:"
echo -e "  ${GREEN}✓${NC} Complete successfully with child workflows"
echo -e "  ${RED}✗${NC} Fail occasionally (payment declined, verification expired, etc.)"
echo -e "  ${YELLOW}⋯${NC} Show multi-step activity progress"
echo ""
echo -e "Without a worker, workflows remain in 'Running' state."
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
