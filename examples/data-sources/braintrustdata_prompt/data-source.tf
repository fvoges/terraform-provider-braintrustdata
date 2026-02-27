# Read a prompt by ID.
data "braintrustdata_prompt" "by_id" {
  # replace with real ID or wire from data/resource
  id = "prompt-abc123"
}

# Read a prompt by name + project context.
data "braintrustdata_prompt" "by_name" {
  name = "support-agent"
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
}

output "prompt_lookup" {
  value = {
    by_id = {
      id            = data.braintrustdata_prompt.by_id.id
      name          = data.braintrustdata_prompt.by_id.name
      slug          = data.braintrustdata_prompt.by_id.slug
      function_type = data.braintrustdata_prompt.by_id.function_type
    }
    by_name = {
      id          = data.braintrustdata_prompt.by_name.id
      created     = data.braintrustdata_prompt.by_name.created
      metadata    = data.braintrustdata_prompt.by_name.metadata
      tags        = data.braintrustdata_prompt.by_name.tags
      description = data.braintrustdata_prompt.by_name.description
    }
  }
}
