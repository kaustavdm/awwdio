/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	darkMode: 'class',
	theme: {
		extend: {
			colors: {
				// Twilio Paste Design System Colors
				twilio: {
					// Primary Blue
					blue: {
						10: '#E1EFFE',
						30: '#6899E8',
						60: '#0263E0', // Primary brand blue
						70: '#0249C0',
						100: '#001C4A'
					},
					// Red (Error/Destructive)
					red: {
						10: '#FCE8E8',
						30: '#F78D8D',
						60: '#D61F1F', // Error red
						70: '#B91919',
						100: '#3B0606'
					},
					// Green (Success)
					green: {
						10: '#D5F3E0',
						30: '#68D68E',
						60: '#14B053', // Success green
						70: '#108F44',
						100: '#04270D'
					},
					// Orange (Warning)
					orange: {
						10: '#FFEAE6',
						30: '#FFA788',
						60: '#FF6B3D',
						70: '#D94E28',
						100: '#4A1609'
					},
					// Purple (Accent)
					purple: {
						10: '#F3EBFA',
						30: '#C99AE8',
						60: '#8957CF',
						70: '#6A3FB2',
						100: '#1F0D36'
					},
					// Gray (Neutral)
					gray: {
						0: '#FFFFFF',
						10: '#F4F4F6',
						20: '#E8EAEA',
						30: '#C4C7C7',
						40: '#A1A6A6',
						50: '#7E8284',
						60: '#5B5F60',
						70: '#3A3C3D',
						80: '#282A2B',
						90: '#1A1C1D',
						100: '#121C2D' // Dark background
					}
				}
			}
		}
	},
	plugins: []
};
