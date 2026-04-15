package backend

import "strings"

// BuildActionPlan generates personalized diet/supplement/exercise/sleep
// recommendations driven by the user's actionable findings.
func BuildActionPlan(summary Summary, categories map[string][]Finding) ActionPlan {
	plan := ActionPlan{}

	// Collect all actionable (non-normal, non-protective) findings
	var actionable []Finding
	for _, findings := range categories {
		for _, f := range findings {
			if f.Status != "normal" && f.Status != "protective" {
				actionable = append(actionable, f)
			}
		}
	}

	for _, f := range actionable {
		gene := strings.ToUpper(f.Gene)
		trait := strings.ToLower(f.Trait)

		// ── DIET ──────────────────────────────────────────────────────
		if containsAny(gene, "MTHFR", "MTR", "MTRR", "BHMT", "PEMT") {
			addUniq(&plan.Diet, "Prioritize natural folate daily: spinach, romaine, arugula, lentils, avocado (2–3 cups dark leafy greens)", f.Gene)
			addUniq(&plan.Diet, "Avoid synthetic folic acid in fortified foods — choose methylated forms only", f.Gene)
		}
		if gene == "PEMT" || containsStr(trait, "choline") {
			addUniq(&plan.Diet, "Increase choline-rich foods: eggs (yolks), liver, salmon, shrimp — target 450–550mg/day", f.Gene)
		}
		if containsAny(gene, "FTO", "MC4R", "LEPR") {
			addUniq(&plan.Diet, "High-protein diet (1.6–2g/kg body weight) at every meal to maximize satiety signaling", f.Gene)
			addUniq(&plan.Diet, "Use smaller plates, pre-portion meals — FTO variants reduce natural satiety cues", f.Gene)
		}
		if containsAny(gene, "FADS1", "FADS2", "ELOVL2") {
			addUniq(&plan.Diet, "Do not rely on plant omega-3 (flaxseed, walnuts) — consume fatty fish 3×/week (salmon, mackerel, sardines)", f.Gene)
		}
		if gene == "AGT" || containsStr(trait, "sodium") || containsStr(trait, "blood pressure") {
			addUniq(&plan.Diet, "Limit sodium to <2,300mg/day, increase potassium (avocado, sweet potato, banana) — target 4,700mg/day", f.Gene)
		}
		if containsAny(gene, "APOE", "CDKN2B-AS1", "LPA") || containsStr(trait, "cardiovascular") || containsStr(trait, "coronary") {
			addUniq(&plan.Diet, "Mediterranean diet pattern: olive oil, fish, legumes, vegetables, nuts — strongest dietary modifier for cardiovascular genes", f.Gene)
		}
		if containsAny(gene, "TCF7L2", "KCNJ11") || containsStr(trait, "diabetes") || containsStr(trait, "insulin") {
			addUniq(&plan.Diet, "Low-glycemic diet: minimize refined carbohydrates and sugar, pair carbs with protein/fat/fiber to blunt glucose spikes", f.Gene)
		}
		if gene == "BCMO1" || containsStr(trait, "vitamin a") || containsStr(trait, "beta-carotene") {
			addUniq(&plan.Diet, "Include preformed vitamin A sources (liver, eggs, full-fat dairy) — cannot rely solely on plant beta-carotene conversion", f.Gene)
		}
		if gene == "APOA5" || containsStr(trait, "triglyceride") {
			addUniq(&plan.Diet, "Minimize refined carbohydrates, alcohol, and sugar — these directly raise triglycerides in APOA5 variant carriers", f.Gene)
		}
		if gene == "APOA2" || containsStr(trait, "saturated fat") {
			addUniq(&plan.Diet, "Limit saturated fat to <7% of total calories — APOA2 variant amplifies saturated fat impact on weight and LDL", f.Gene)
		}
		if gene == "HFE" || containsStr(trait, "iron") || containsStr(trait, "hemochromatosis") {
			addUniq(&plan.Diet, "Avoid iron-fortified foods and cast-iron cooking; limit red meat to 1–2×/week; avoid vitamin C with iron-containing meals", f.Gene)
		}
		if containsStr(trait, "lactose") || gene == "LCT" {
			addUniq(&plan.Diet, "Limit or eliminate dairy; choose lactase-enzyme-treated products or fermented dairy (kefir, aged cheese) if needed", f.Gene)
		}
		if gene == "ALDH2" || containsStr(trait, "alcohol") || containsStr(trait, "acetaldehyde") {
			addUniq(&plan.Diet, "Minimize or eliminate alcohol — ALDH2 variant causes toxic acetaldehyde accumulation and significantly raises esophageal cancer risk", f.Gene)
		}

		// ── SUPPLEMENTS ───────────────────────────────────────────────
		if containsAny(gene, "MTHFR", "MTR", "MTRR") {
			addUniq(&plan.Supplements, "Methylfolate (5-MTHF) 400–800mcg/day — active form bypassing impaired MTHFR conversion", f.Gene)
			addUniq(&plan.Supplements, "Methylcobalamin (B12) 500–1,000mcg/day sublingual — synergistic with 5-MTHF for homocysteine clearance", f.Gene)
		}
		if containsAny(gene, "VDR", "CYP2R1", "DHCR7") || containsStr(trait, "vitamin d") {
			addUniq(&plan.Supplements, "Vitamin D3 3,000–5,000 IU/day + K2 (MK-7) 100mcg — VDR variants require higher D3 for adequate tissue response", f.Gene)
		}
		if containsAny(gene, "FADS1", "FADS2", "ELOVL2") || containsStr(trait, "omega-3") {
			addUniq(&plan.Supplements, "Omega-3 fish oil or algal oil 2–3g EPA+DHA/day — plant omega-3 conversion severely impaired", f.Gene)
		}
		if gene == "COMT" || containsStr(trait, "dopamine") || containsStr(trait, "catecholamine") {
			addUniq(&plan.Supplements, "Magnesium glycinate 300–400mg at night — COMT enzyme cofactor, supports sleep and stress resilience", f.Gene)
			addUniq(&plan.Supplements, "Rhodiola rosea 300–500mg/day — adaptogen balancing dopamine/norepinephrine without overstimulation", f.Gene)
		}
		if containsAny(gene, "ACTN3", "PPARGC1A") || containsStr(trait, "mitochondri") || containsStr(trait, "endurance") {
			addUniq(&plan.Supplements, "CoQ10 ubiquinol 100–200mg/day with fat-containing meal — supports mitochondrial efficiency", f.Gene)
		}
		if containsAny(gene, "SLC30A8") || containsStr(trait, "zinc") || containsStr(trait, "beta cell") {
			addUniq(&plan.Supplements, "Zinc picolinate 15–25mg/day with food — supports insulin secretion, VDR function, and immune health", f.Gene)
		}
		if gene == "APOE" || containsStr(trait, "alzheimer") || containsStr(trait, "cognitive") {
			addUniq(&plan.Supplements, "DHA 1–2g/day (algal or fish oil) — direct brain omega-3, most evidence-based cognitive supplement for APOE variants", f.Gene)
			addUniq(&plan.Supplements, "Lion's Mane mushroom 500–1,000mg/day — stimulates nerve growth factor (NGF), supports neurogenesis", f.Gene)
		}
		if gene == "SLCO1B1" {
			addUniq(&plan.Supplements, "CoQ10 100–200mg/day if on statin therapy — SLCO1B1 variant increases statin-induced myopathy risk", f.Gene)
		}
		if containsAny(gene, "TNF", "IL6") || containsStr(trait, "inflammation") {
			addUniq(&plan.Supplements, "Omega-3 2–3g/day + curcumin 500mg with piperine — directly suppresses TNF-alpha and IL-6 inflammatory cytokines", f.Gene)
		}
		if gene == "TCN2" || gene == "FUT2" || containsStr(trait, "b12") {
			addUniq(&plan.Supplements, "Methylcobalamin or adenosylcobalamin 1,000mcg/day — standard cyanocobalamin not effective for TCN2/FUT2 variants", f.Gene)
		}
		if containsStr(trait, "glutathione") || gene == "CBS" || gene == "GSTM1" || gene == "GSTT1" {
			addUniq(&plan.Supplements, "NAC (N-acetylcysteine) 600–1,200mg/day — glutathione precursor, critical for transsulfuration pathway variants", f.Gene)
		}

		// ── EXERCISE ──────────────────────────────────────────────────
		if gene == "ACTN3" {
			if f.A1 == "T" && f.A2 == "T" {
				addUniq(&plan.Exercise, "Prioritize Zone 2 aerobic training (60–70% max HR, 45–60 min, 4–5×/week) — ACTN3 XX is endurance-dominant by design", f.Gene)
			} else if f.A1 == "C" && f.A2 == "C" {
				addUniq(&plan.Exercise, "Heavy compound lifts (3–6 rep range) and sprint/explosive training — ACTN3 RR is power-dominant, responds best to high intensity", f.Gene)
			} else {
				addUniq(&plan.Exercise, "Balanced training: mix aerobic with moderate-heavy resistance — ACTN3 RX responds well to both modalities", f.Gene)
			}
		}
		if containsAny(gene, "FTO", "TCF7L2", "KCNJ11") {
			addUniq(&plan.Exercise, "Consistent aerobic exercise 5+ days/week — directly counteracts FTO/T2D genetic risk, reduces phenotype expression 50–60%", f.Gene)
		}
		if containsAny(gene, "PPARGC1A", "UCP1", "UCP2") || containsStr(trait, "mitochondri") {
			addUniq(&plan.Exercise, "Allow longer adaptation periods (8–12 weeks) before assessing endurance gains — mitochondrial biogenesis variants respond more slowly", f.Gene)
		}
		if containsAny(gene, "NOS3", "ADRB2") || containsStr(trait, "nitric oxide") || containsStr(trait, "vasodilation") {
			addUniq(&plan.Exercise, "Dietary nitrate pre-exercise (beetroot juice, spinach, arugula) — boosts NO-mediated vasodilation independent of NOS3 variants", f.Gene)
		}
		if gene == "APOE" || containsStr(trait, "alzheimer") || containsStr(trait, "cognitive") {
			addUniq(&plan.Exercise, "Daily aerobic exercise (30+ min, any intensity) — most powerful intervention for APOE ε4 risk reduction; promotes amyloid clearance", f.Gene)
		}
		if containsAny(gene, "COL1A1", "COL5A1") || containsStr(trait, "tendon") || containsStr(trait, "collagen") {
			addUniq(&plan.Exercise, "Thorough warm-up and cool-down every session; avoid abrupt training volume spikes — collagen gene variants increase tendon/ligament injury risk", f.Gene)
		}

		// ── SLEEP ─────────────────────────────────────────────────────
		if gene == "CLOCK" || containsStr(trait, "chronotype") || containsStr(trait, "circadian") {
			addUniq(&plan.Sleep, "Keep consistent sleep/wake times within ±1 hour, 7 days/week — circadian gene variants are highly sensitive to schedule drift", f.Gene)
		}
		if containsAny(gene, "BTBD9", "MEIS1") || containsStr(trait, "restless legs") {
			addUniq(&plan.Sleep, "Check ferritin levels — target >75 ng/mL (iron deficiency is the strongest modifiable restless legs trigger)", f.Gene)
		}
		if gene == "MTNR1B" || containsStr(trait, "melatonin") {
			addUniq(&plan.Sleep, "Finish dinner by 7pm — MTNR1B variants impair beta-cell glucose tolerance when eating close to melatonin onset", f.Gene)
		}
		if gene == "COMT" || containsStr(trait, "dopamine") || containsStr(trait, "stress") {
			addUniq(&plan.Sleep, "No screens after 9pm, dim lighting by 8pm — COMT variants leave catecholamines elevated longer, delaying sleep onset under evening stimulation", f.Gene)
		}
		if gene == "ADA" || containsStr(trait, "slow wave") || containsStr(trait, "deep sleep") {
			addUniq(&plan.Sleep, "Avoid caffeine after 1pm — ADA variants increase slow-wave sleep intensity, and caffeine-related adenosine buildup is amplified", f.Gene)
		}

		// ── MONITORING ────────────────────────────────────────────────
		if gene == "HFE" || containsStr(trait, "iron") || containsStr(trait, "hemochromatosis") {
			addUniq(&plan.Monitoring, "Annual ferritin + transferrin saturation blood test — catch iron accumulation early before organ damage", f.Gene)
		}
		if containsAny(gene, "APOE", "CDKN2B-AS1") || containsStr(trait, "cardiovascular") || containsStr(trait, "coronary") {
			addUniq(&plan.Monitoring, "Annual lipid panel (LDL, HDL, triglycerides, ApoB) starting at age 30 — aggressive prevention for cardiac risk gene carriers", f.Gene)
		}
		if gene == "APOE" || containsStr(trait, "alzheimer") {
			addUniq(&plan.Monitoring, "Annual cognitive screening after age 40 (MoCA or SAGE) — early detection window for APOE ε4 carriers", f.Gene)
		}
		if containsAny(gene, "TCF7L2", "KCNJ11", "PPARG") || containsStr(trait, "diabetes") {
			addUniq(&plan.Monitoring, "Annual fasting glucose + HbA1c — T2D risk gene carriers benefit from early lifestyle intervention before prediabetes progresses", f.Gene)
		}
		if containsAny(gene, "BRCA1", "BRCA2") || containsStr(trait, "breast cancer") || containsStr(trait, "ovarian") {
			addUniq(&plan.Monitoring, "Annual mammogram + consider breast MRI; discuss RRSO timing with gynecologic oncologist — BRCA carriers have specific surveillance protocols", f.Gene)
		}
		if gene == "MTHFR" || containsStr(trait, "homocysteine") {
			addUniq(&plan.Monitoring, "Annual homocysteine blood test — target <8 µmol/L; elevated homocysteine is a direct cardiovascular and cognitive risk factor", f.Gene)
		}
		if containsAny(gene, "VDR", "CYP2R1") || containsStr(trait, "vitamin d") {
			addUniq(&plan.Monitoring, "25-OH Vitamin D blood test every 6 months — VDR variants require higher circulating D3 levels; target 50–80 ng/mL", f.Gene)
		}
		if gene == "LPA" || containsStr(trait, "lipoprotein") {
			addUniq(&plan.Monitoring, "Lp(a) blood test — if elevated (>50 mg/dL), discuss with cardiologist; it is largely genetically determined and not modified by diet", f.Gene)
		}
	}

	// Universal defaults if categories are sparse
	if len(plan.Diet) == 0 {
		plan.Diet = append(plan.Diet, ActionItem{Text: "Whole-food, predominantly plant-based diet with adequate protein (1.4–2g/kg body weight)", Gene: "General"})
	}
	if len(plan.Supplements) == 0 {
		plan.Supplements = append(plan.Supplements, ActionItem{Text: "Vitamin D3 2,000 IU + K2 100mcg daily — widely deficient and foundational for most health pathways", Gene: "General"})
	}
	if len(plan.Exercise) == 0 {
		plan.Exercise = append(plan.Exercise, ActionItem{Text: "150+ minutes moderate aerobic exercise weekly + 2–3 resistance training sessions — baseline for all genetic profiles", Gene: "General"})
	}
	if len(plan.Sleep) == 0 {
		plan.Sleep = append(plan.Sleep, ActionItem{Text: "7–9 hours consistent nightly sleep — most powerful free intervention for hormone balance, cognition, and cardiovascular health", Gene: "General"})
	}
	if len(plan.Monitoring) == 0 {
		plan.Monitoring = append(plan.Monitoring, ActionItem{Text: "Annual comprehensive metabolic panel + CBC — baseline bloodwork for tracking health trends over time", Gene: "General"})
	}

	// Always add resistance training as a universal recommendation
	addUniq(&plan.Exercise, "Resistance training 3×/week — builds metabolically active muscle, the strongest long-term modifier of metabolic genetic risk", "General")
	addUniq(&plan.Sleep, "Cool, dark room at 65–68°F (18–20°C) with blackout curtains — temperature and light are universal sleep quality drivers", "General")
	addUniq(&plan.Sleep, "Target 7–9 hours nightly — sleep is when amyloid clearance, hormone restoration, and DNA repair occur", "General")

	return plan
}

// addUniq appends an ActionItem if the text does not already exist in the slice.
func addUniq(items *[]ActionItem, text, gene string) {
	for _, existing := range *items {
		if existing.Text == text {
			return
		}
	}
	*items = append(*items, ActionItem{Text: text, Gene: gene})
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if s == sub {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}
