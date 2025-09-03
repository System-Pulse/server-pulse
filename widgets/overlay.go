package widgets

import (
	"fmt"
	"strings"
	// "github.com/charmbracelet/lipgloss"
)

// Fonction helper pour créer un overlay manuellement
func createOverlay(baseContent, overlayContent string, width, height, overlayWidth, overlayHeight int) string {
	// Diviser le contenu de base en lignes
	baseLines := strings.Split(baseContent, "\n")

	// Diviser l'overlay en lignes
	overlayLines := strings.Split(overlayContent, "\n")

	// Calculer la position du menu (centré)
	posX := (width - overlayWidth) / 2
	posY := (height - overlayHeight) / 3

	// S'assurer que posY est dans les limites
	if posY < 0 {
		posY = 0
	}
	if posY+len(overlayLines) > len(baseLines) {
		posY = len(baseLines) - len(overlayLines)
		if posY < 0 {
			posY = 0
		}
	}

	// Créer les lignes de l'overlay avec fond opaque
	var resultLines []string

	for y := 0; y < len(baseLines); y++ {
		if y >= posY && y < posY+len(overlayLines) {
			// Ligne où le menu doit apparaître
			overlayIndex := y - posY
			if overlayIndex < len(overlayLines) {
				// Créer une ligne avec fond opaque
				opaqueLine := createOpaqueLine(baseLines[y], overlayLines[overlayIndex], posX, overlayWidth, width)
				resultLines = append(resultLines, opaqueLine)
			} else {
				resultLines = append(resultLines, baseLines[y])
			}
		} else {
			// Ligne normale
			resultLines = append(resultLines, baseLines[y])
		}

		// Limiter le nombre de lignes à la hauteur de l'écran
		if y >= height-1 {
			break
		}
	}

	return strings.Join(resultLines, "\n")
}

func createOpaqueLine(baseLine, overlayLine string, posX, overlayWidth, totalWidth int) string {
	// Créer une ligne complète avec fond opaque pour la zone du menu
	if len(baseLine) < totalWidth {
		baseLine += strings.Repeat(" ", totalWidth-len(baseLine))
	}

	// Convertir la ligne en runes pour manipulation précise
	baseRunes := []rune(baseLine)
	overlayRunes := []rune(overlayLine)

	// S'assurer que la position est dans les limites
	if posX < 0 {
		posX = 0
	}
	if posX+len(overlayRunes) > len(baseRunes) {
		// Ajuster la longueur de l'overlay si nécessaire
		overlayRunes = overlayRunes[:len(baseRunes)-posX]
	}

	// Créer la ligne résultante avec fond opaque seulement pour la zone du menu
	resultRunes := make([]rune, len(baseRunes))
	copy(resultRunes, baseRunes)

	// Appliquer l'overlay avec fond opaque
	for i := 0; i < len(overlayRunes); i++ {
		if posX+i < len(resultRunes) {
			resultRunes[posX+i] = overlayRunes[i]
		}
	}

	return string(resultRunes)
}

// Alternative plus simple : créer un rectangle opaque complet pour le menu
func createOpaqueOverlay(baseContent, overlayContent string, width, height int) string {
	baseLines := strings.Split(baseContent, "\n")
	overlayLines := strings.Split(overlayContent, "\n")

	// Calculer la position centrée
	// menuWidth := lipgloss.Width(overlayContent)
	menuHeight := len(overlayLines)
	// posX := (width - menuWidth) / 2
	posY := (height - menuHeight) / 3

	var resultLines []string

	for y := 0; y < len(baseLines); y++ {
		if y >= posY && y < posY+menuHeight {
			overlayIndex := y - posY
			if overlayIndex < len(overlayLines) {
				// Pour les lignes du menu, utiliser uniquement le contenu du menu avec fond opaque
				resultLines = append(resultLines, overlayLines[overlayIndex])
			} else {
				resultLines = append(resultLines, baseLines[y])
			}
		} else {
			resultLines = append(resultLines, baseLines[y])
		}
	}

	return strings.Join(resultLines, "\n")
}

func createSimpleOverlay(baseContent, overlayContent string, width, height int) string {
	baseLines := strings.Split(baseContent, "\n")
	overlayLines := strings.Split(overlayContent, "\n")

	menuHeight := len(overlayLines)
	menuWidth := 0
	for _, line := range overlayLines {
		if len(line) > menuWidth {
			menuWidth = len(line)
		}
	}

	posY := (height - menuHeight) / 3
	if posY < 0 {
		posY = 0
	}

	var resultLines []string

	for y, line := range baseLines {
		if y >= posY && y < posY+menuHeight && (y-posY) < len(overlayLines) {
			// Ligne avec menu
			menuLine := overlayLines[y-posY]
			// Centrer horizontalement
			padding := (width - len(menuLine)) / 2
			if padding < 0 {
				padding = 0
			}
			centeredLine := strings.Repeat(" ", padding) + menuLine
			resultLines = append(resultLines, centeredLine)
		} else {
			// Ligne normale (rendue semi-transparente)
			if len(line) > 0 {
				// Raccourcir la ligne pour indiquer qu'elle est en arrière-plan
				resultLines = append(resultLines, line[:min(len(line), width/2)]+"...")
			} else {
				resultLines = append(resultLines, line)
			}
		}
	}

	return strings.Join(resultLines, "\n")
}

func insertOverlayLine(baseLine, overlayLine string, posX, overlayWidth int) string {
	// S'assurer que la ligne de base est assez longue
	if len(baseLine) < posX+len(overlayLine) {
		// Ajouter des espaces si nécessaire
		baseLine += strings.Repeat(" ", posX+len(overlayLine)-len(baseLine))
	}

	// Remplacer la partie de la ligne
	if len(baseLine) > posX+len(overlayLine) {
		return baseLine[:posX] + overlayLine + baseLine[posX+len(overlayLine):]
	} else {
		return baseLine[:posX] + overlayLine
	}
}

func createANSIOverlay(baseContent, overlayContent string, width, height int) string {
	overlayLines := strings.Split(overlayContent, "\n")
	menuHeight := len(overlayLines)
	menuWidth := 0
	for _, line := range overlayLines {
		if len(line) > menuWidth {
			menuWidth = len(line)
		}
	}

	posY := (height - menuHeight) / 3
	posX := (width - menuWidth) / 2

	// ANSI escape codes pour sauvegarder la position, déplacer le curseur, et restaurer
	ansiPrefix := "\x1b7" // Sauvegarder la position du curseur
	ansiSuffix := "\x1b8" // Restaurer la position du curseur

	var overlay strings.Builder
	overlay.WriteString(baseContent)
	overlay.WriteString(ansiPrefix)

	for i, line := range overlayLines {
		// Déplacer le curseur à la position du menu
		overlay.WriteString(fmt.Sprintf("\x1b[%d;%dH", posY+i+1, posX+1))
		// Fond noir opaque
		overlay.WriteString("\x1b[48;5;0m")
		overlay.WriteString("\x1b[38;5;15m") // Texte blanc
		overlay.WriteString(line)
		overlay.WriteString("\x1b[0m") // Reset
	}

	overlay.WriteString(ansiSuffix)
	return overlay.String()
}
