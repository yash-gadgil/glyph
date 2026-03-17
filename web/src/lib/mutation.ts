import { useMutation } from "@tanstack/react-query";

export default function mutationBuilder() {
  return function mutation() {
    return useMutation({});
  };
}
