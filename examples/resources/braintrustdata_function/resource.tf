resource "braintrustdata_project" "ai_functions" {
  name        = "function-example-project"
  description = "Project for function resource examples"
}

# Prompt-backed function with metadata and tags.
resource "braintrustdata_function" "support_tool" {
  project_id    = braintrustdata_project.ai_functions.id
  name          = "support-tool"
  slug          = "support-tool"
  description   = "Prompt-backed support tool"

  function_data = jsonencode({
    type = "prompt"
  })

  prompt_data = jsonencode({
    prompt = {
      type = "chat"
      messages = [
        {
          role    = "system"
          content = "You are a helpful support assistant."
        }
      ]
    }
    options = {
      model = "gpt-4o-mini"
    }
  })

  metadata = {
    owner = "ml-platform"
    tier  = "production"
  }

  tags = ["support", "production"]
}

output "function_id" {
  value = braintrustdata_function.support_tool.id
}
