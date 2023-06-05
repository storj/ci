set -euo pipefail

# Requirements for UI tests
npx playwright install-deps
npx playwright install
npm install -g @playwright/test
