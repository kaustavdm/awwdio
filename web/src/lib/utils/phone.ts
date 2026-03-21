/**
 * Obfuscate a phone number for display.
 * +15551234567 -> +1***1234567 (show country code prefix + last 4)
 */
export function obfuscatePhone(phone: string): string {
  const hasPlus = phone.startsWith('+');
  const digits = phone.replace(/\D/g, '');
  if (digits.length < 7) return phone;

  const last4 = digits.slice(-4);
  const prefix = digits.slice(0, Math.max(1, digits.length - 7));
  const masked = prefix + '***' + last4;
  return (hasPlus ? '+' : '') + masked;
}

/**
 * Check if a participant identity looks like a phone number.
 * Phone identities from Twilio start with + and are all digits.
 */
export function isPhoneIdentity(identity: string): boolean {
  return /^\+?\d{7,15}$/.test(identity.replace(/[\s()\-]/g, ''));
}
