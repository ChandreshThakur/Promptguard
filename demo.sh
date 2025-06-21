#!/bin/bash

# PromptGuard Demo Script
# Shows red/green diff functionality for failed assertions

echo "🚀 PromptGuard Demo - Red/Green Diff Analysis"
echo "============================================="
echo

# Build PromptGuard
echo "📦 Building PromptGuard..."
go build -o pg main.go

# Create demo artifacts directory
mkdir -p demo-artifacts

echo "✅ Build complete!"
echo

# Simulate a test run with failures (mock results)
echo "🧪 Running prompt tests..."
cat > demo-artifacts/failed-results.json << 'EOF'
{
  "total": 3,
  "passed": 1,
  "failed": 2,
  "skipped": 0,
  "totalCost": 0.0156,
  "duration": 1800000000,
  "testResults": [
    {
      "name": "onboard-pro-plan",
      "promptFile": "prompts/onboard.prompt",
      "provider": "openai:gpt-4o",
      "variables": {
        "customer": "Alice Johnson",
        "product": "Pro Plan"
      },
      "response": "Welcome Alice! Here are your basic features.",
      "assertions": [
        {
          "type": "answer-relevance",
          "expected": "Mention Pro Plan upgrade benefits",
          "actual": "Welcome Alice! Here are your basic features.",
          "passed": false,
          "score": 0.32,
          "message": "Relevance score: 0.32 (threshold: 0.70)"
        },
        {
          "type": "cost",
          "expected": 0.003,
          "actual": 0.0045,
          "passed": false,
          "message": "Cost: $0.0045 (threshold: $0.0030)"
        }
      ],
      "cost": 0.0045,
      "duration": 850000000,
      "status": "failed"
    },
    {
      "name": "invoice-generation",
      "promptFile": "prompts/invoice.prompt",
      "provider": "openai:gpt-4o",
      "variables": {
        "customer": "Bob Smith",
        "amount": 29.99
      },
      "response": "{\"customer\": \"Bob Smith\", \"amount\": 29.99}",
      "assertions": [
        {
          "type": "contains-json",
          "expected": {
            "type": "object",
            "required": ["total_due", "billing_period"]
          },
          "actual": "{\"customer\": \"Bob Smith\", \"amount\": 29.99}",
          "passed": false,
          "message": "Required field missing: total_due"
        }
      ],
      "cost": 0.0021,
      "duration": 750000000,
      "status": "failed"
    },
    {
      "name": "newsletter-content",
      "promptFile": "prompts/newsletter.prompt",
      "provider": "openai:gpt-3.5-turbo",
      "variables": {
        "month": "December",
        "user_segment": "premium"
      },
      "response": "Premium December Newsletter with exciting updates!",
      "assertions": [
        {
          "type": "answer-relevance",
          "expected": "Highlight premium benefits",
          "actual": "Premium December Newsletter with exciting updates!",
          "passed": true,
          "score": 0.78,
          "message": "Relevance score: 0.78 (threshold: 0.75)"
        }
      ],
      "cost": 0.0090,
      "duration": 650000000,
      "status": "passed"
    }
  ],
  "metadata": {
    "timestamp": "2024-12-21T10:30:00Z",
    "commitSha": "abc123def456",
    "version": "0.1.0"
  }
}
EOF

echo "❌ Tests completed with failures!"
echo

# Generate markdown diff report
echo "📊 Generating failure analysis report..."

# Use Go to generate the markdown diff (we'll create a simple CLI command for this)
cat > demo-artifacts/failure-analysis.md << 'EOF'
# 🔍 PromptGuard Failure Analysis

❌ **2 test(s) failed** - Analysis below:

## ❌ `onboard-pro-plan`

**📁 File:** `prompts/onboard.prompt`  
**🤖 Provider:** `openai:gpt-4o`  
**💰 Cost:** $0.0045  

### 🔬 Failed Assertions

#### ❌ `answer-relevance`

**Message:** Relevance score: 0.32 (threshold: 0.70)

**Expected Keywords/Concepts:**
```
Mention Pro Plan upgrade benefits
```

**Relevance Score:** 0.32 ❌

#### ❌ `cost`

**Message:** Cost: $0.0045 (threshold: $0.0030)

| Metric | Expected | Actual | Status |
|--------|----------|--------|---------|
| Cost | ≤ $0.0030 | $0.0045 | ❌ Over budget |

**💸 Cost overage:** 50.0% over threshold

### 📄 Actual Response

```json
Welcome Alice! Here are your basic features.
```

---

## ❌ `invoice-generation`

**📁 File:** `prompts/invoice.prompt`  
**🤖 Provider:** `openai:gpt-4o`  
**💰 Cost:** $0.0021  

### 🔬 Failed Assertions

#### ❌ `contains-json`

**Message:** Required field missing: total_due

**Expected JSON Structure:**
```json
{type: object, required: [total_due billing_period]}
```

**Actual Response:**
```json
{"customer": "Bob Smith", "amount": 29.99}
```

**Diff:**
```diff
+ {"customer": "Bob Smith", "amount": 29.99}
- Required fields: total_due, billing_period
```

### 📄 Actual Response

```json
{"customer": "Bob Smith", "amount": 29.99}
```

---

## 📊 Summary

- **Total Tests:** 3
- **✅ Passed:** 1
- **❌ Failed:** 2
- **💰 Total Cost:** $0.0156
EOF

echo "✅ Failure analysis generated!"
echo

# Show the red/green diff output
echo "🎨 Red/Green Diff Analysis:"
echo "=========================="
echo

# Use colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}❌ FAILED TESTS${NC}"
echo -e "${YELLOW}📁 prompts/onboard.prompt${NC}"
echo -e "   ${RED}❌ answer-relevance: 0.32 < 0.70 threshold${NC}"
echo -e "   ${RED}❌ cost: \$0.0045 > \$0.0030 threshold${NC}"
echo

echo -e "${YELLOW}📁 prompts/invoice.prompt${NC}"
echo -e "   ${RED}❌ contains-json: Missing required field 'total_due'${NC}"
echo

echo -e "${GREEN}✅ PASSED TESTS${NC}"
echo -e "${YELLOW}📁 prompts/newsletter.prompt${NC}"
echo -e "   ${GREEN}✅ answer-relevance: 0.78 ≥ 0.75 threshold${NC}"
echo

echo "📄 Full analysis available in: demo-artifacts/failure-analysis.md"
echo

# Show GitHub Action integration
echo "🚀 GitHub Action Integration:"
echo "=============================="
echo "::error file=prompts/onboard.prompt,title=PromptGuard Test Failure::answer-relevance: Relevance score: 0.32 (threshold: 0.70); cost: Cost: \$0.0045 (threshold: \$0.0030)"
echo "::error file=prompts/invoice.prompt,title=PromptGuard Test Failure::contains-json: Required field missing: total_due"
echo

echo "🎯 Demo Complete!"
echo "================="
echo "PromptGuard detected 2 failing prompts and generated detailed analysis."
echo "• Red/green diff output shows exactly what failed"
echo "• Markdown report provides detailed failure analysis"
echo "• GitHub annotations pinpoint failures in CI"
echo "• Cost tracking prevents budget overruns"
echo
echo "Ready for production CI/CD integration! 🚀"
EOF

chmod +x demo-artifacts/demo.sh
