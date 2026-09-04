export function randomString(len = 8): string {
  return Math.random().toString(36).substring(2, 2 + len);
}

export function randomEmail(): string {
  return `test_${randomString()}@e2e.com`;
}

export function randomWorkspace(): { name: string; slug: string } {
  const id = randomString(6);
  return { name: `E2E Workspace ${id}`, slug: `e2e-${id}` };
}

export function randomProject(): { name: string; identifier: string; description: string } {
  const id = randomString(4).toUpperCase();
  return { name: `E2E Project ${id}`, identifier: id, description: `Auto-generated project ${id}` };
}
