import { Menu } from "lucide-react";
import { Expandable, ExpandableTrigger } from "../primitives/Expandable";

export default function ExpandableNav() {

  return (

    <Expandable
      expandDirection="both"
      expandBehavior="replace"
      initialDelay={0.2}
    >

      {({ isExpanded }: { isExpanded: boolean }) => (
        <ExpandableTrigger>
          {!isExpanded && (
            <Menu className="w-6 h-6 md:hidden pointer-events-auto" />
          )}
        </ExpandableTrigger>
      )}
    </Expandable>
  );
}