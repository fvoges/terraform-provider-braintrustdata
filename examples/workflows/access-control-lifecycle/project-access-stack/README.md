# Project Access Stack

This stack owns projects, role groups, ACLs, and role-group memberships.

## Run

```bash
cd examples/workflows/access-control-lifecycle/project-access-stack
terraform init
terraform apply
```

This stack reads `role_group_member_group_ids_by_binding_key` from `../teams-stack/terraform.tfstate`.

## Apply order

1. Apply `teams-stack`
2. Apply `project-access-stack`
3. Re-apply only the stack whose source model changed
