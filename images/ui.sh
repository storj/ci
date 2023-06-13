set -euo pipefail

# Requirements for UI tests
# npx playwright install-deps
# npx playwright install
npx playwright install --with-deps
npm install -g @playwright/test
