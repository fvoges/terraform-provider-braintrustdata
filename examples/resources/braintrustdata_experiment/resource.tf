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

# Create a simple experiment
resource "braintrustdata_experiment" "simple" {
  name       = "gpt-4-baseline"
  project_id = "proj-abc123"
}

# Create an experiment with description
resource "braintrustdata_experiment" "with_description" {
  name        = "prompt-optimization-v1"
  project_id  = "proj-abc123"
  description = "Testing different prompt variations for customer support responses"
}

# Create an experiment with metadata and tags
resource "braintrustdata_experiment" "with_metadata" {
  name       = "model-comparison"
  project_id = "proj-abc123"
  metadata = {
    version     = "1.0"
    model       = "gpt-4"
    temperature = "0.7"
    dataset     = "customer-support-v2"
  }
  tags = ["production", "customer-support", "gpt-4"]
}

# Create a public experiment
resource "braintrustdata_experiment" "public" {
  name        = "open-research-experiment"
  project_id  = "proj-abc123"
  description = "Public experiment for research purposes"
  public      = true
  tags        = ["research", "public"]
}

# Create an experiment with all optional attributes
resource "braintrustdata_experiment" "complete" {
  name        = "full-featured-experiment"
  project_id  = "proj-abc123"
  description = "Comprehensive experiment with all configuration options"
  public      = false
  metadata = {
    model_family    = "gpt-4"
    use_case        = "summarization"
    evaluation_type = "human-preference"
    cost_per_run    = "0.05"
  }
  tags = ["ml-ops", "summarization", "cost-optimized"]
  repo_info = {
    commit         = "abc123def456"
    branch         = "main"
    tag            = "v1.2.3"
    dirty          = false
    author_name    = "Jane Developer"
    author_email   = "jane@example.com"
    commit_message = "Tune summarization prompt"
    commit_time    = "2026-02-18T12:00:00Z"
  }
}
