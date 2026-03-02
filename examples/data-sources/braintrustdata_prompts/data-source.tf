# List prompts in a project.
data "braintrustdata_prompts" "all" {
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
}

# Filter prompts by name and slug.
data "braintrustdata_prompts" "filtered" {
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
  name       = "support-agent"
  slug       = "support-agent"
}

output "prompt_lists" {
  value = {
    all_ids      = data.braintrustdata_prompts.all.ids
    filtered_ids = data.braintrustdata_prompts.filtered.ids
  }
}
