-- Remove ON DELETE CASCADE from issue-related foreign keys to support soft delete
-- When an issue is soft-deleted, its related data should NOT be physically deleted

-- Comments
ALTER TABLE public.comments DROP CONSTRAINT IF EXISTS fk_comments_issue;
ALTER TABLE public.comments ADD CONSTRAINT fk_comments_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

-- Issue Custom Field Values
ALTER TABLE public.issue_custom_field_values DROP CONSTRAINT IF EXISTS fk_issue_custom_field_values_issue;
ALTER TABLE public.issue_custom_field_values ADD CONSTRAINT fk_issue_custom_field_values_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

-- Issue Relations
ALTER TABLE public.issue_relations DROP CONSTRAINT IF EXISTS fk_issue_relations_issue;
ALTER TABLE public.issue_relations ADD CONSTRAINT fk_issue_relations_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

ALTER TABLE public.issue_relations DROP CONSTRAINT IF EXISTS fk_issue_relations_related_issue;
ALTER TABLE public.issue_relations ADD CONSTRAINT fk_issue_relations_related_issue FOREIGN KEY (related_issue_id) REFERENCES public.issues(id);

-- Module Issues
ALTER TABLE public.module_issues DROP CONSTRAINT IF EXISTS fk_module_issues_issue;
ALTER TABLE public.module_issues ADD CONSTRAINT fk_module_issues_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

-- Issue Assignees (created by GORM AutoMigrate)
ALTER TABLE public.issue_assignees DROP CONSTRAINT IF EXISTS fk_issue_assignees_issue;
ALTER TABLE public.issue_assignees ADD CONSTRAINT fk_issue_assignees_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

-- Issue Labels (created by GORM AutoMigrate)
ALTER TABLE public.issue_labels DROP CONSTRAINT IF EXISTS fk_issue_labels_issue;
ALTER TABLE public.issue_labels ADD CONSTRAINT fk_issue_labels_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

-- Issue Cycles (created by GORM AutoMigrate)
ALTER TABLE public.issue_cycles DROP CONSTRAINT IF EXISTS fk_issue_cycles_issue;
ALTER TABLE public.issue_cycles ADD CONSTRAINT fk_issue_cycles_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);

-- Issue Pages (created by GORM AutoMigrate)
ALTER TABLE public.issue_pages DROP CONSTRAINT IF EXISTS fk_issue_pages_issue;
ALTER TABLE public.issue_pages ADD CONSTRAINT fk_issue_pages_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);