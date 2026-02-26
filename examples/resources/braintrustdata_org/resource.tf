# Manage the current organization (explicit org_id shown for clarity).
resource "braintrustdata_org" "current" {
  org_id = "org-***" # replace with real organization ID, or rely on provider organization_id

  # Optional organization settings can be managed when needed.
  image_rendering_mode = "click_to_load"
}

output "organization" {
  value = {
    id                  = braintrustdata_org.current.id
    name                = braintrustdata_org.current.name
    image_rendering_mode = braintrustdata_org.current.image_rendering_mode
  }
}
