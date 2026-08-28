# Supabase on Railway
Deploy your own self-hosted instance of Supabase on Railway.

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/supabase?referralCode=IQhE0B&utm_medium=integration&utm_source=template&utm_campaign=generic)

## Deployment Guide

- Deploy the template above, filling the Auth URL with your application URL (or a placeholder, like `http://localhost:3000`).
- After reading the entire description of the `IMPORTANT_READ_ME` variable, type anything (this value doesn't matter).
- Choose a method below to generate your project's secrets:
  - Deploy the template as-is. After the project has been created, use https://6ixfalls.github.io/supabase/. Use `Connect Railway` to connect to the newly created project and click generate and deploy. This will automatically seed your database secrets.
  - Fill in the variables in the deployment interface. Using https://6ixfalls.github.io/supabase/, generate the credentials and scroll down to `Railway variables`. Copy the value of each variable and paste it accordingly.
- After deployment, Supabase should be completely up and running.

> [!NOTE]
> PostgreSQL is intentionally built from this repository's `postgres/Dockerfile` instead of deployed directly from the upstream Supabase image. Keep the repository-backed Postgres service so Railway's database tabs remain available.
> Repository-backed services use the repository root as their Docker build context. Keep Railway's Root Directory at the repository root and select the matching service Dockerfile.

## Known Issues
- Railway's deployment order is not reliable. During deployment, the database may come up after other services (Like Supabase Auth), which will show a "schema not initialized" error. Simply redeploy the image in the Railway menu under the running replica > Deploy, or Command+K and select `Redeploy source image`.

## Best-effort Maintenance

This repository & template are maintained on a best-effort basis; major issues should be reported under [issues](https://github.com/6ixfalls/supabase/issues?q=sort%3Aupdated-desc+is%3Aissue+state%3Aopen+). Check [Supabase Self-Hosted Changelog](https://github.com/supabase/supabase/blob/master/docker/CHANGELOG.md) for important updates. We try our best to keep in sync with them.

## Deployment Maintenance
Your deployments do not automatically update; you will have to do so manually, following the self-hosted Supabase guidance and the changes made within https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml (and other related files). You can use the latest template version and compare it to your deployed version to see what has changed.

The official guidance for self-hosted instances applies, but many (if not all) scripts will not work with Railway. You'll need to adapt the workflow and manually apply the migrations to upgrade major Postgres versions, like in https://supabase.com/docs/guides/self-hosting/postgres-upgrade-17.
