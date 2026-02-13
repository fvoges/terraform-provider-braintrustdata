terraform {
  required_providers {
    braintrustdata = {
      source = "braintrustdata/braintrustdata"
    }
  }
}

# Configure the provider with API credentials
# API key can be set via BRAINTRUST_API_KEY environment variable
provider "braintrustdata" {
  # api_key = "sk-***" # Set via BRAINTRUST_API_KEY environment variable
}

# Create a simple dataset
resource "braintrustdata_dataset" "simple" {
  name       = "customer-support-v1"
  project_id = "proj-abc123"
}

# Create a dataset with description
resource "braintrustdata_dataset" "with_description" {
  name        = "evaluation-dataset-v2"
  project_id  = "proj-abc123"
  description = "Curated dataset for evaluating customer support responses"
}

# Create a dataset with metadata and tags
resource "braintrustdata_dataset" "with_metadata" {
  name       = "training-dataset"
  project_id = "proj-abc123"
  metadata = {
    version      = "1.0"
    source       = "production-logs"
    sample_count = "10000"
    date_range   = "2024-01-01-to-2024-03-31"
  }
  tags = ["production", "training", "q1-2024"]
}

# Create a dataset with all optional attributes
resource "braintrustdata_dataset" "complete" {
  name        = "full-featured-dataset"
  project_id  = "proj-abc123"
  description = "Comprehensive dataset with all configuration options"
  metadata = {
    data_type       = "conversation"
    use_case        = "summarization"
    quality_score   = "0.95"
    annotation_type = "human-labeled"
  }
  tags = ["ml-ops", "summarization", "high-quality"]
}
