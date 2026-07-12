import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const mobileCss = readFileSync(resolve(__dirname, "../mobile.css"), "utf8");

const normalizedCss = mobileCss.replace(/\s+/g, " ");

describe("mobile file listing styles", () => {
  it("hides size column entirely and modified column for rows on mobile", () => {
    expect(normalizedCss).toContain(
      "#listing.list .item .size { display: none;"
    );
    expect(normalizedCss).toContain(
      "#listing.list .item:not(.header) .modified { display: none;"
    );
  });

  it("hides both size and modified headers on very small screens", () => {
    expect(normalizedCss).toContain(
      "#listing.list .item.header .modified { display: none;"
    );
  });
});
