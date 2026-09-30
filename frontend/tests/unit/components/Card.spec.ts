import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Card from "@/components/Card.vue";

describe("Card", () => {
  it("applies card-body when padded is true (default)", () => {
    const w = mount(Card, { slots: { default: "x" } });
    expect(w.classes()).toContain("card");
    expect(w.classes()).toContain("card-body");
  });

  it("omits card-body when padded is false", () => {
    const w = mount(Card, { props: { padded: false }, slots: { default: "x" } });
    expect(w.classes()).toContain("card");
    expect(w.classes()).not.toContain("card-body");
  });

  it("renders default slot", () => {
    const w = mount(Card, { slots: { default: "<p>hello</p>" } });
    expect(w.text()).toBe("hello");
  });
});