import { DialogApiInjection } from 'naive-ui/es/dialog/src/DialogProvider';
import { MessageApiInjection } from 'naive-ui/es/message/src/MessageProvider';
interface NaiveUiType {
    message: MessageApiInjection;
    dialog: DialogApiInjection;
}
declare const naiveui: NaiveUiType;
export default naiveui;
