package ovpn

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type VpnProfile struct {
	ID          string   `json:"id"`
	Filename    string   `json:"filename"`
	CountryCode string   `json:"country_code"`
	Country     string   `json:"country"`
	City        string   `json:"city"`
	IsVirtual   bool     `json:"is_virtual"`
	Endpoints   []string `json:"endpoints"`
	FullPath    string   `json:"full_path"`
}

var countryCodeMap = map[string]string{
	"AD": "Andorra",
	"AE": "United Arab Emirates",
	"AL": "Albania",
	"AM": "Armenia",
	"AR": "Argentina",
	"AT": "Austria",
	"AU": "Australia",
	"AZ": "Azerbaijan",
	"BA": "Bosnia and Herzegovina",
	"BD": "Bangladesh",
	"BE": "Belgium",
	"BG": "Bulgaria",
	"BM": "Bermuda",
	"BN": "Brunei",
	"BO": "Bolivia",
	"BR": "Brazil",
	"BS": "Bahamas",
	"BT": "Bhutan",
	"BZ": "Belize",
	"CA": "Canada",
	"CH": "Switzerland",
	"CL": "Chile",
	"CO": "Colombia",
	"CR": "Costa Rica",
	"CY": "Cyprus",
	"CZ": "Czech Republic",
	"DE": "Germany",
	"DK": "Denmark",
	"DO": "Dominican Republic",
	"DZ": "Algeria",
	"EC": "Ecuador",
	"EE": "Estonia",
	"EG": "Egypt",
	"ES": "Spain",
	"FI": "Finland",
	"FR": "France",
	"GE": "Georgia",
	"GH": "Ghana",
	"GR": "Greece",
	"GT": "Guatemala",
	"HK": "Hong Kong",
	"HN": "Honduras",
	"HR": "Croatia",
	"HT": "Haiti",
	"HU": "Hungary",
	"ID": "Indonesia",
	"IE": "Ireland",
	"IL": "Israel",
	"IM": "Isle of Man",
	"IN": "India",
	"IQ": "Iraq",
	"IR": "Iran",
	"IS": "Iceland",
	"IT": "Italy",
	"JE": "Jersey",
	"JM": "Jamaica",
	"JO": "Jordan",
	"JP": "Japan",
	"KE": "Kenya",
	"KG": "Kyrgyzstan",
	"KH": "Cambodia",
	"KR": "South Korea",
	"KW": "Kuwait",
	"KY": "Cayman Islands",
	"KZ": "Kazakhstan",
	"LA": "Laos",
	"LB": "Lebanon",
	"LI": "Liechtenstein",
	"LK": "Sri Lanka",
	"LT": "Lithuania",
	"LU": "Luxembourg",
	"LV": "Latvia",
	"MA": "Morocco",
	"MC": "Monaco",
	"MD": "Moldova",
	"ME": "Montenegro",
	"MK": "North Macedonia",
	"ML": "Mali",
	"MM": "Myanmar",
	"MN": "Mongolia",
	"MO": "Macau",
	"MT": "Malta",
	"MU": "Mauritius",
	"MV": "Maldives",
	"MX": "Mexico",
	"MY": "Malaysia",
	"NG": "Nigeria",
	"NI": "Nicaragua",
	"NL": "Netherlands",
	"NO": "Norway",
	"NP": "Nepal",
	"NZ": "New Zealand",
	"OM": "Oman",
	"PA": "Panama",
	"PE": "Peru",
	"PG": "Papua New Guinea",
	"PH": "Philippines",
	"PK": "Pakistan",
	"PL": "Poland",
	"PR": "Puerto Rico",
	"PT": "Portugal",
	"PY": "Paraguay",
	"QA": "Qatar",
	"RO": "Romania",
	"RS": "Serbia",
	"SA": "Saudi Arabia",
	"SE": "Sweden",
	"SG": "Singapore",
	"SI": "Slovenia",
	"SK": "Slovakia",
	"SN": "Senegal",
	"TH": "Thailand",
	"TR": "Turkey",
	"TT": "Trinidad and Tobago",
	"TW": "Taiwan",
	"TZ": "Tanzania",
	"UA": "Ukraine",
	"UG": "Uganda",
	"UK": "United Kingdom",
	"GB": "United Kingdom",
	"US": "United States",
	"UY": "Uruguay",
	"UZ": "Uzbekistan",
	"VE": "Venezuela",
	"VN": "Vietnam",
	"ZA": "South Africa",
	"ZM": "Zambia",
	"ZW": "Zimbabwe",
}

func CountryCodeToName(code string) string {
	if name, ok := countryCodeMap[strings.ToUpper(code)]; ok {
		return name
	}
	return "Unknown"
}

func ParseProfileFromContent(filename, content, fullPath string) (*VpnProfile, error) {
	stem := strings.TrimSuffix(filename, ".ovpn")
	id := stem
	isVirtual := strings.Contains(stem, "Virtual")

	countryCode := "UN"
	city := stem

	parts := strings.Split(stem, "-")
	if len(parts) >= 3 && parts[0] == "NCVPN" {
		countryCode = strings.ToUpper(strings.TrimSpace(parts[1]))
		var cityParts []string
		for _, p := range parts[2:] {
			trimmed := strings.TrimSpace(p)
			if trimmed == "UDP" || trimmed == "Virtual" {
				continue
			}
			cityParts = append(cityParts, trimmed)
		}
		if len(cityParts) > 0 {
			city = strings.Join(cityParts, "-")
		}
	}

	countryName := CountryCodeToName(countryCode)
	if countryName == "Unknown" && city != "" {
		countryName = city
	}

	var endpoints []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "remote ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				endpoints = append(endpoints, fmt.Sprintf("%s:%s", fields[1], fields[2]))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed scanning ovpn profile %q: %w", filename, err)
	}

	return &VpnProfile{
		ID:          id,
		Filename:    filename,
		CountryCode: countryCode,
		Country:     countryName,
		City:        city,
		IsVirtual:   isVirtual,
		Endpoints:   endpoints,
		FullPath:    fullPath,
	}, nil
}

func LoadProfiles(dir string) ([]VpnProfile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles dir %q: %w", dir, err)
	}

	var profiles []VpnProfile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ovpn") {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		prof, err := ParseProfileFromContent(entry.Name(), string(data), fullPath)
		if err == nil && prof != nil {
			profiles = append(profiles, *prof)
		}
	}

	sort.Slice(profiles, func(i, j int) bool {
		c1 := strings.ToLower(profiles[i].Country)
		c2 := strings.ToLower(profiles[j].Country)
		if c1 != c2 {
			return c1 < c2
		}
		city1 := strings.ToLower(profiles[i].City)
		city2 := strings.ToLower(profiles[j].City)
		if city1 != city2 {
			return city1 < city2
		}
		return strings.ToLower(profiles[i].ID) < strings.ToLower(profiles[j].ID)
	})

	return profiles, nil
}


