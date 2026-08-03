# Supabase on Railway
Deploy your own self-hosted instance of Supabase on Railway.

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/supabase?referralCode=IQhE0B&utm_medium=integration&utm_source=template&utm_campaign=generic)

## Best-effort Maintenance

This repository & template are maintained on a best-effort basis; major issues should be reported under [issues](https://github.com/6ixfalls/supabase/issues?q=sort%3Aupdated-desc+is%3Aissue+state%3Aopen+). Check [Supabase Self-Hosted Changelog](https://github.com/supabase/supabase/blob/master/docker/CHANGELOG.md) for important updates. We try our best to keep in sync with them.

## Deployment Maintenance
Your deployments do not automatically update; you will have to do so manually, following the self-hosted Supabase guidance and the changes made within https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml (and other related files). You can use the latest template version and compare it to your deployed version to see what has changed.

The official guidance for self-hosted instances applies, but many (if not all) scripts will not work with Railway. You'll need to adapt the workflow and manually apply the migrations to upgrade major Postgres versions, like in https://supabase.com/docs/guides/self-hosting/postgres-upgrade-17.

