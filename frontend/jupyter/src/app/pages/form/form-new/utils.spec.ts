import { calculateLimits, configSizeToNumber } from './utils';

describe('form-new utils', () => {
  describe('configSizeToNumber', () => {
    it('returns NaN for null', () => {
      expect(configSizeToNumber(null)).toBeNaN();
    });

    it('returns NaN for undefined', () => {
      expect(configSizeToNumber(undefined)).toBeNaN();
    });

    it('returns a number input unchanged', () => {
      expect(configSizeToNumber(5)).toBe(5);
    });

    it('strips the Gi suffix from a size string', () => {
      expect(configSizeToNumber('16Gi')).toBe(16);
    });

    it('parses a fractional Gi size string', () => {
      expect(configSizeToNumber('2.5Gi')).toBe(2.5);
    });

    it('parses a plain numeric string', () => {
      expect(configSizeToNumber('16')).toBe(16);
    });

    it('returns NaN for a non-numeric string', () => {
      expect(configSizeToNumber('abc')).toBeNaN();
    });
  });

  describe('calculateLimits', () => {
    it('multiplies two numbers and formats to one decimal', () => {
      expect(calculateLimits(2, 3)).toBe('6.0');
    });

    it('handles a Gi size string as the request', () => {
      expect(calculateLimits('4Gi', 2)).toBe('8.0');
    });

    it('handles size strings for both arguments', () => {
      expect(calculateLimits('4Gi', '1.5')).toBe('6.0');
    });

    it('returns null when the request is null', () => {
      expect(calculateLimits(null, 2)).toBeNull();
    });

    it('returns null when an argument is not numeric', () => {
      expect(calculateLimits('abc', 2)).toBeNull();
    });
  });
});
