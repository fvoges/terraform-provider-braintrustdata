resource "braintrustdata_project" "ai_assistant" {
  name        = "prompt-example-project"
  description = "Project for prompt examples"
}

# Minimal prompt with just a name.
resource "braintrustdata_prompt" "minimal" {
  name       = "support-agent"
  project_id = braintrustdata_project.ai_assistant.id
}

# Prompt with description, tags, and metadata.
resource "braintrustdata_prompt" "customer_support" {
  name        = "customer-support-v1"
  project_id  = braintrustdata_project.ai_assistant.id
  description = "Customer support assistant prompt"

  metadata = {
    model       = "gpt-4"
    temperature = "0.7"
    team        = "ml"
  }

  tags = ["customer-support", "production"]
}

# Prompt with structured prompt_data using jsonencode.
resource "braintrustdata_prompt" "structured" {
  name        = "code-reviewer"
  project_id  = braintrustdata_project.ai_assistant.id
  description = "Code review assistant"

  prompt_data = jsonencode({
    prompt = {
      type    = "chat"
      messages = [
        {
          role    = "system"
          content = "You are an expert code reviewer. Provide concise, actionable feedback."
        }
      ]
    }
    options = {
      model       = "gpt-4"
      temperature = 0.3
      max_tokens  = 2048
    }
  })

  tags = ["code-review", "engineering"]
}

output "prompt_ids" {
  value = {
    minimal          = braintrustdata_prompt.minimal.id
    customer_support = braintrustdata_prompt.customer_support.id
    structured       = braintrustdata_prompt.structured.id
  }
}
