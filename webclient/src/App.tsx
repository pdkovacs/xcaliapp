import { isEqual } from "lodash";
import { useEffect, useMemo, useRef, useState } from "react";
import { LinearProgress } from "@mui/material";
import { Excalidraw, MainMenu } from "@excalidraw/excalidraw";
import type { NonDeletedExcalidrawElement, Ordered } from "@excalidraw/excalidraw/element/types";
import { OpenDrawingDialog } from "./features/drawing/OpenDrawing";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types";
import { useAppDispatch, useAppSelector } from "./app/hooks";
import { selectSavedDrawing, selectDrawingToEditStatus, selectCurrentDrawingContent, drawingContentChanged } from "./features/drawing/drawingSlice";
import { SaveDrawingDialog } from "./features/drawing/SaveDrawing";

import "@excalidraw/excalidraw/index.css";

import "./App.css";

export type XcalidrawDocument = Readonly<{
  type: string;
  title: string;
  elements: readonly Ordered<NonDeletedExcalidrawElement>[];
}>


const App = () => {

  const [excalidrawAPI, setExcalidrawAPI] = useState<ExcalidrawImperativeAPI | null>(null);
  const excalidrawAPIUnsubscribe = useRef<(() => void) | null>(null);

  const currentDrawingStatus = useAppSelector(selectDrawingToEditStatus);
  const savedDrawing = useAppSelector(selectSavedDrawing);
  const currentContent = useAppSelector(selectCurrentDrawingContent);

  const [openDrawingDialogOpen, setOpenDrawingDialogOpen] = useState(false);
  const [saveDrawingDialogOpen, setSaveDrawingDialogOpen] = useState(false);

  const dispatch = useAppDispatch();

  useEffect(() => {
    if (excalidrawAPI) {
      if (excalidrawAPIUnsubscribe.current !== null) {
        excalidrawAPIUnsubscribe.current();
      }
      excalidrawAPIUnsubscribe.current = excalidrawAPI.onChange(() => {
        const excalidrawContent = excalidrawAPI.getSceneElements();
        const doc: XcalidrawDocument = {
          type: "excalidraw",
          title: "asdfasdffas asdfasdf",
          elements: excalidrawContent
        };
        dispatch(drawingContentChanged(JSON.stringify(doc)));
      });
    }
  }, [excalidrawAPI, savedDrawing]);

  useEffect(() => {
    const sceneData = {
      elements: savedDrawing.content ? JSON.parse(savedDrawing.content).elements : [],
      appState: {}
    };

    excalidrawAPI?.updateScene(sceneData);

    document.title = savedDrawing.title;
  }, [savedDrawing]);

  const contentHasChanged = useMemo(() => {
    return !isEqual(savedDrawing.content, currentContent);
  }, [savedDrawing, currentContent]);

  console.log(">>>>>>>> contentHasChanged", contentHasChanged);

  return (
    <div>
      <div className="document-title">{savedDrawing.title}</div>
      <div>
        {currentDrawingStatus === "loading" && <LinearProgress sx={{ marginTop: "-4px" }} />
        }
        <div className="xcali-area">
          <Excalidraw excalidrawAPI={api => setExcalidrawAPI(api)}>
            <MainMenu>
              <MainMenu.Item onSelect={() => setOpenDrawingDialogOpen(true)}>
                Open
              </MainMenu.Item>
              <MainMenu.Item disabled={!contentHasChanged} onSelect={() => setSaveDrawingDialogOpen(true)}>
                Save
              </MainMenu.Item>
            </MainMenu>
          </Excalidraw>
        </div>
        <OpenDrawingDialog open={openDrawingDialogOpen} onClose={() => setOpenDrawingDialogOpen(false)} />
        <SaveDrawingDialog open={saveDrawingDialogOpen} onClose={() => setSaveDrawingDialogOpen(false)} />
      </div>
    </div>
  );
};

export default App;
